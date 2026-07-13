package execution

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type LeaseRepository interface {
	Claim(context.Context, string, time.Time, time.Duration, int) (Execution, bool, error)
	Renew(context.Context, string, string, time.Time, time.Duration) (bool, error)
	Complete(context.Context, string, string, ExecutionResponse, time.Time) error
}

type Sandbox interface {
	Execute(context.Context, ExecutionSpec) (ExecutionResponse, error)
}

type WorkerOptions struct {
	Owner                string
	MaxActionTimeout     time.Duration
	SandboxResponseGrace time.Duration
	RPCDeadline          time.Duration
	PersistenceGrace     time.Duration
	LeaseDuration        time.Duration
	HeartbeatInterval    time.Duration
	PollInterval         time.Duration
	MaxClaims            int
	Now                  func() time.Time
}

type Worker struct {
	repository LeaseRepository
	sandbox    Sandbox
	options    WorkerOptions
}

func NewWorker(repository LeaseRepository, sandbox Sandbox, options WorkerOptions) (*Worker, error) {
	if repository == nil || sandbox == nil {
		return nil, fmt.Errorf("execution lease repository and Sandbox client are required")
	}
	if options.Owner == "" || options.MaxActionTimeout <= 0 || options.SandboxResponseGrace <= 0 ||
		options.RPCDeadline <= 0 || options.PersistenceGrace <= 0 || options.LeaseDuration <= 0 ||
		options.HeartbeatInterval <= 0 || options.PollInterval <= 0 || options.MaxClaims < 1 {
		return nil, fmt.Errorf("execution worker timing, owner, and max claims must be positive")
	}
	if options.MaxActionTimeout+options.SandboxResponseGrace >= options.RPCDeadline {
		return nil, fmt.Errorf("RPC deadline must exceed max action timeout plus Sandbox response grace")
	}
	if options.RPCDeadline+options.PersistenceGrace >= options.LeaseDuration {
		return nil, fmt.Errorf("worker lease must exceed RPC deadline plus persistence grace")
	}
	if options.HeartbeatInterval >= options.LeaseDuration/2 {
		return nil, fmt.Errorf("worker heartbeat must be less than half the lease duration")
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Worker{repository: repository, sandbox: sandbox, options: options}, nil
}

func (w *Worker) Run(ctx context.Context) error {
	for {
		processed, err := w.RunOnce(ctx)
		if err != nil && !errors.Is(err, ErrLeaseLost) {
			return err
		}
		if processed {
			continue
		}
		timer := time.NewTimer(w.options.PollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (w *Worker) RunOnce(ctx context.Context) (bool, error) {
	claimed, ok, err := w.repository.Claim(ctx, w.options.Owner, w.options.Now().UTC(), w.options.LeaseDuration, w.options.MaxClaims)
	if err != nil || !ok {
		return false, err
	}
	started := w.options.Now().UTC()
	rpcContext, cancel := context.WithTimeout(ctx, w.options.RPCDeadline)
	defer cancel()
	type sandboxResult struct {
		response ExecutionResponse
		err      error
	}
	resultChannel := make(chan sandboxResult, 1)
	go func() {
		response, err := w.sandbox.Execute(rpcContext, claimed.Spec)
		resultChannel <- sandboxResult{response: response, err: err}
	}()
	heartbeat := time.NewTicker(w.options.HeartbeatInterval)
	defer heartbeat.Stop()

	var result sandboxResult
	for {
		select {
		case <-ctx.Done():
			cancel()
			return true, ctx.Err()
		case <-heartbeat.C:
			renewed, err := w.repository.Renew(ctx, claimed.ID, w.options.Owner, w.options.Now().UTC(), w.options.LeaseDuration)
			if err != nil {
				cancel()
				return true, err
			}
			if !renewed {
				cancel()
				return true, ErrLeaseLost
			}
		case result = <-resultChannel:
			cancel()
			goto completed
		}
	}

completed:
	response := result.response
	if result.err != nil {
		code, message := "sandbox_unreachable", "Sandbox request failed before producing a result"
		if errors.Is(result.err, ErrInvalidSandboxResponse) {
			code, message = "invalid_sandbox_response", "Sandbox returned an invalid execution response"
		} else if errors.Is(rpcContext.Err(), context.DeadlineExceeded) {
			code, message = "sandbox_rpc_deadline", "Sandbox did not respond before the RPC deadline"
		}
		response = workerInfraResponse(claimed.Spec, started, w.options.Now().UTC(), code, message)
	} else if response.ExecutionID != claimed.ID {
		response = workerInfraResponse(claimed.Spec, started, w.options.Now().UTC(), "invalid_sandbox_response", "Sandbox response identity did not match the queued execution")
	} else if err := response.Validate(); err != nil {
		response = workerInfraResponse(claimed.Spec, started, w.options.Now().UTC(), "invalid_sandbox_response", "Sandbox returned an invalid execution response")
	}
	if err := w.repository.Complete(ctx, claimed.ID, w.options.Owner, response, w.options.Now().UTC()); err != nil {
		return true, err
	}
	return true, nil
}

func workerInfraResponse(spec ExecutionSpec, started, finished time.Time, code, message string) ExecutionResponse {
	duration := finished.Sub(started).Milliseconds()
	if duration < 0 {
		duration = 0
	}
	return ExecutionResponse{
		ProtocolVersion: ProtocolVersion, ExecutionID: spec.ExecutionID,
		Status: ExecutionInfraFailed, Stages: []StageResult{}, DurationMS: duration,
		Policy:  PolicyReport{Network: NetworkPolicyReport{Requested: spec.Policy.Network, Enforcement: EnforcementPolicyOnly}},
		Failure: &Failure{Code: code, Message: message},
	}
}
