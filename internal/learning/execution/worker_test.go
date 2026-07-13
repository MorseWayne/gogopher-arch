package execution

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWorkerCompletesSandboxResultUnderCurrentLease(t *testing.T) {
	spec := validSpec(ActionBuild)
	repository := &leaseRepositoryStub{claimed: Execution{ID: spec.ExecutionID, Spec: spec}}
	worker, err := NewWorker(repository, sandboxStub{response: successResponse(spec)}, workerTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	processed, err := worker.RunOnce(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if !processed || repository.completed.Status != ExecutionSucceeded || repository.completedID != spec.ExecutionID {
		t.Fatalf("processed=%v completed=%#v", processed, repository.completed)
	}
}

func TestWorkerClassifiesSandboxErrorsAndDeadlineAsInfrastructure(t *testing.T) {
	for _, test := range []struct {
		name string
		box  Sandbox
		code string
	}{
		{name: "unreachable", box: sandboxStub{err: errors.New("connection refused")}, code: "sandbox_unreachable"},
		{name: "invalid response", box: sandboxStub{err: ErrInvalidSandboxResponse}, code: "invalid_sandbox_response"},
		{name: "deadline", box: blockingSandbox{}, code: "sandbox_rpc_deadline"},
	} {
		t.Run(test.name, func(t *testing.T) {
			spec := validSpec(ActionTest)
			repository := &leaseRepositoryStub{claimed: Execution{ID: spec.ExecutionID, Spec: spec}, renewResult: true}
			options := workerTestOptions()
			options.RPCDeadline = 40 * time.Millisecond
			options.LeaseDuration = 100 * time.Millisecond
			options.HeartbeatInterval = 20 * time.Millisecond
			worker, err := NewWorker(repository, test.box, options)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := worker.RunOnce(t.Context()); err != nil {
				t.Fatal(err)
			}
			if repository.completed.Status != ExecutionInfraFailed || repository.completed.Failure == nil || repository.completed.Failure.Code != test.code {
				t.Fatalf("completed = %#v", repository.completed)
			}
		})
	}
}

func TestWorkerDoesNotWriteTerminalStateAfterLeaseLoss(t *testing.T) {
	spec := validSpec(ActionTest)
	repository := &leaseRepositoryStub{claimed: Execution{ID: spec.ExecutionID, Spec: spec}, renewResult: false}
	options := workerTestOptions()
	options.HeartbeatInterval = 10 * time.Millisecond
	worker, err := NewWorker(repository, blockingSandbox{}, options)
	if err != nil {
		t.Fatal(err)
	}
	processed, err := worker.RunOnce(t.Context())
	if !processed || !errors.Is(err, ErrLeaseLost) || repository.completeCalls != 0 {
		t.Fatalf("processed=%v error=%v complete calls=%d", processed, err, repository.completeCalls)
	}
}

func TestWorkerInterruptionLeavesExecutionForLeaseRecovery(t *testing.T) {
	spec := validSpec(ActionTest)
	repository := &leaseRepositoryStub{claimed: Execution{ID: spec.ExecutionID, Spec: spec}, renewResult: true}
	started := make(chan struct{})
	worker, err := NewWorker(repository, signalingSandbox{started: started}, workerTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	result := make(chan error, 1)
	go func() {
		_, err := worker.RunOnce(ctx)
		result <- err
	}()
	<-started
	cancel()
	if err := <-result; !errors.Is(err, context.Canceled) || repository.completeCalls != 0 {
		t.Fatalf("RunOnce() error=%v complete calls=%d", err, repository.completeCalls)
	}
}

func TestWorkerRejectsAmbiguousTimeoutOrdering(t *testing.T) {
	options := workerTestOptions()
	options.RPCDeadline = options.MaxActionTimeout + options.SandboxResponseGrace
	if _, err := NewWorker(&leaseRepositoryStub{}, sandboxStub{}, options); err == nil {
		t.Fatal("NewWorker(ambiguous timeout) error = nil")
	}
	options = workerTestOptions()
	options.LeaseDuration = options.RPCDeadline + options.PersistenceGrace
	if _, err := NewWorker(&leaseRepositoryStub{}, sandboxStub{}, options); err == nil {
		t.Fatal("NewWorker(ambiguous lease) error = nil")
	}
}

type leaseRepositoryStub struct {
	claimed       Execution
	claimResult   bool
	renewResult   bool
	completedID   string
	completed     ExecutionResponse
	completeCalls int
}

func (s *leaseRepositoryStub) Claim(context.Context, string, time.Time, time.Duration, int) (Execution, bool, error) {
	if s.claimed.ID == "" {
		return Execution{}, s.claimResult, nil
	}
	return s.claimed, true, nil
}

func (s *leaseRepositoryStub) Renew(context.Context, string, string, time.Time, time.Duration) (bool, error) {
	return s.renewResult, nil
}

func (s *leaseRepositoryStub) Complete(_ context.Context, id, _ string, response ExecutionResponse, _ time.Time) error {
	s.completedID = id
	s.completed = response
	s.completeCalls++
	return nil
}

type sandboxStub struct {
	response ExecutionResponse
	err      error
}

func (s sandboxStub) Execute(context.Context, ExecutionSpec) (ExecutionResponse, error) {
	return s.response, s.err
}

type blockingSandbox struct{}

func (blockingSandbox) Execute(ctx context.Context, _ ExecutionSpec) (ExecutionResponse, error) {
	<-ctx.Done()
	return ExecutionResponse{}, ctx.Err()
}

type signalingSandbox struct{ started chan<- struct{} }

func (s signalingSandbox) Execute(ctx context.Context, _ ExecutionSpec) (ExecutionResponse, error) {
	close(s.started)
	<-ctx.Done()
	return ExecutionResponse{}, ctx.Err()
}

func workerTestOptions() WorkerOptions {
	return WorkerOptions{
		Owner: "worker-1", MaxActionTimeout: 10 * time.Millisecond, SandboxResponseGrace: 10 * time.Millisecond,
		RPCDeadline: 50 * time.Millisecond, PersistenceGrace: 10 * time.Millisecond,
		LeaseDuration: 100 * time.Millisecond, HeartbeatInterval: 20 * time.Millisecond,
		PollInterval: 10 * time.Millisecond, MaxClaims: 2, Now: time.Now,
	}
}
