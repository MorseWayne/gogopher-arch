package evaluation

import (
	"strconv"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

func TestM201LearningLoopSolutionsPassRealReleaseAndSandboxEvaluation(t *testing.T) {
	tests := []struct {
		name        string
		activityID  string
		solution    map[string]string
		explanation string
	}{
		{
			name:       "guided request slice",
			activityID: "practice-http-request-slice",
			solution: map[string]string{
				"httpslice/handler.go": httpRequestSliceSolution(),
			},
			explanation: "我先让 request ID middleware 派生 context，再交给 ServeMux 选择 handler，并用 httptest 从请求到响应验证整条链路。",
		},
		{
			name:       "independent gocheck hub server",
			activityID: "assessment-gocheck-http",
			solution: map[string]string{
				"httpserver/server.go":      gocheckHTTPServerSolution(),
				"httpserver/server_test.go": gocheckHTTPTableTestSolution(),
			},
			explanation: "一次请求先由 request ID middleware 基于原 request 的 context 派生新值，再交给 ServeMux 按 method 和 path 选择 handler，handler 只调用注入依赖并写 response。Server 的四类 timeout 分别约束 header、读取、写入和空闲连接；取消到达后用独立 deadline 调用 Shutdown，先关闭 listener 与空闲连接并等待活动请求结束，再回收 Serve 返回的 ErrServerClosed，因此函数不会在请求尚未完成时提前退出。",
		},
		{
			name:       "delayed jobwatch variant",
			activityID: "review-jobwatch-http",
			solution: map[string]string{
				"adminserver/server.go":      jobwatchHTTPServerSolution(),
				"adminserver/server_test.go": jobwatchHTTPTableTestSolution(),
			},
			explanation: "jobwatch 的请求先进入 trace middleware，把 trace ID 放进新的 context 并写响应头，然后由 ServeMux 区分 readiness 与 job handler。http.Server 的 ReadHeader、Read、Write 和 Idle timeout 都由配置显式提供；停止时 Shutdown 关闭 listener、清理空闲连接并等待活动 handler 返回，之后才读取 Serve 的 ErrServerClosed 并作为正常关闭收敛，所以不会丢掉仍在执行的请求。",
		},
	}

	registry := draftReleaseRegistry(t)
	for index, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			activity, err := registry.ActivityView(registry.CurrentReleaseID(), testCase.activityID, 1)
			if err != nil {
				t.Fatal(err)
			}
			task, err := registry.ExecutionTask(registry.CurrentReleaseID(), activity.TaskRef.ID, activity.TaskRef.Version)
			if err != nil {
				t.Fatal(err)
			}
			workspace, err := registry.PublicWorkspace(registry.CurrentReleaseID(), task.ID, task.Version)
			if err != nil {
				t.Fatal(err)
			}
			for path, source := range testCase.solution {
				workspace[path] = source
			}
			current := attempt.Attempt{
				ID: "00000000-0000-4000-9000-00000000000" + strconv.Itoa(index+1), ReleaseID: registry.CurrentReleaseID(),
				ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash,
				TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
				Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace),
			}
			builder, err := execution.NewSpecBuilder(registry)
			if err != nil {
				t.Fatal(err)
			}
			executionID := "00000000-0000-4000-9100-00000000000" + strconv.Itoa(index+1)
			specification, err := builder.Build(current, executionID, execution.ActionSubmit)
			if err != nil {
				t.Fatal(err)
			}
			response, err := runRegressionSandbox(t, specification)
			if err != nil {
				t.Fatal(err)
			}
			if response.Status != execution.ExecutionSucceeded {
				t.Fatalf("sandbox response = %#v", response)
			}
			frozen := submission.Submission{
				ID: "00000000-0000-4000-9200-00000000000" + strconv.Itoa(index+1), AttemptID: current.ID,
				ReleaseID: current.ReleaseID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
				Workspace: workspace, Explanation: testCase.explanation,
			}
			terminal := execution.Execution{
				ID: executionID, AttemptID: current.ID, SubmissionID: frozen.ID,
				TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
				Status: response.Status, Response: &response,
			}
			generator, err := NewGenerator(registry)
			if err != nil {
				t.Fatal(err)
			}
			results, err := generator.Generate(frozen, terminal)
			if err != nil {
				t.Fatal(err)
			}
			for _, result := range results {
				if result.Status != execution.RulePassed {
					t.Fatalf("rule %s = %#v", result.RuleID, result)
				}
			}
		})
	}
}

func httpRequestSliceSolution() string {
	return `package httpslice

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type TargetLookup func(context.Context, string) (string, bool)
type RequestIDGenerator func() string
type requestIDKey struct{}

func NewHandler(lookup TargetLookup, nextRequestID RequestIDGenerator) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /targets/{id}", func(w http.ResponseWriter, r *http.Request) {
		name, found := lookup(r.Context(), r.PathValue("id"))
		if !found {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, name)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if id == "" {
			id = strings.TrimSpace(nextRequestID())
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}
`
}

func gocheckHTTPServerSolution() string {
	return `package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type TargetLookup func(context.Context, string) (string, bool)
type Dependencies struct {
	LookupTarget TargetLookup
	NextRequestID func() string
}
type Timeouts struct {
	ReadHeader time.Duration
	Read time.Duration
	Write time.Duration
	Idle time.Duration
}
type requestIDKey struct{}

func NewHandler(dependencies Dependencies) (http.Handler, error) {
	if dependencies.LookupTarget == nil || dependencies.NextRequestID == nil {
		return nil, errors.New("all handler dependencies are required")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /targets/{id}", func(w http.ResponseWriter, r *http.Request) {
		name, found := dependencies.LookupTarget(r.Context(), r.PathValue("id"))
		if !found {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, name)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if id == "" {
			id = strings.TrimSpace(dependencies.NextRequestID())
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		mux.ServeHTTP(w, r.WithContext(ctx))
	}), nil
}

func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

func NewServer(handler http.Handler, timeouts Timeouts) (*http.Server, error) {
	if handler == nil || timeouts.ReadHeader <= 0 || timeouts.Read <= 0 || timeouts.Write <= 0 || timeouts.Idle <= 0 {
		return nil, errors.New("handler and positive timeouts are required")
	}
	return &http.Server{
		Handler: handler,
		ReadHeaderTimeout: timeouts.ReadHeader,
		ReadTimeout: timeouts.Read,
		WriteTimeout: timeouts.Write,
		IdleTimeout: timeouts.Idle,
	}, nil
}

func Serve(ctx context.Context, server *http.Server, listener net.Listener, shutdownTimeout time.Duration) error {
	if ctx == nil || server == nil || listener == nil || shutdownTimeout <= 0 {
		return errors.New("context, server, listener and positive shutdown timeout are required")
	}
	serveErrors := make(chan error, 1)
	go func() { serveErrors <- server.Serve(listener) }()
	select {
	case err := <-serveErrors:
		return err
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		shutdownErr := server.Shutdown(shutdownContext)
		if shutdownErr != nil {
			_ = server.Close()
		}
		serveErr := <-serveErrors
		if shutdownErr != nil {
			return shutdownErr
		}
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			return serveErr
		}
		return nil
	}
}
`
}

func gocheckHTTPTableTestSolution() string {
	return `package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerContract(t *testing.T) {
	handler, err := NewHandler(Dependencies{
		LookupTarget: func(_ context.Context, id string) (string, bool) { return id, id == "api" },
		NextRequestID: func() string { return "test-request" },
	})
	if err != nil { t.Fatal(err) }
	tests := []struct {
		name string
		method string
		path string
		want int
	}{
		{name: "health", method: http.MethodGet, path: "/healthz", want: http.StatusNoContent},
		{name: "target", method: http.MethodGet, path: "/targets/api", want: http.StatusOK},
		{name: "missing", method: http.MethodGet, path: "/targets/missing", want: http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(testCase.method, testCase.path, nil))
			if recorder.Code != testCase.want { t.Fatalf("status = %d", recorder.Code) }
		})
	}
}
`
}

func jobwatchHTTPServerSolution() string {
	return `package adminserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type JobLookup func(context.Context, string) (string, bool)
type Dependencies struct {
	Ready func(context.Context) bool
	LookupJob JobLookup
	NextTraceID func() string
}
type Timeouts struct {
	ReadHeader time.Duration
	Read time.Duration
	Write time.Duration
	Idle time.Duration
}
type traceIDKey struct{}

func NewHandler(dependencies Dependencies) (http.Handler, error) {
	if dependencies.Ready == nil || dependencies.LookupJob == nil || dependencies.NextTraceID == nil {
		return nil, errors.New("all handler dependencies are required")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if !dependencies.Ready(r.Context()) {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		status, found := dependencies.LookupJob(r.Context(), r.PathValue("id"))
		if !found {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, status)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.Header.Get("X-Trace-ID"))
		if id == "" {
			id = strings.TrimSpace(dependencies.NextTraceID())
		}
		w.Header().Set("X-Trace-ID", id)
		ctx := context.WithValue(r.Context(), traceIDKey{}, id)
		mux.ServeHTTP(w, r.WithContext(ctx))
	}), nil
}

func TraceID(ctx context.Context) string {
	id, _ := ctx.Value(traceIDKey{}).(string)
	return id
}

func NewServer(handler http.Handler, timeouts Timeouts) (*http.Server, error) {
	if handler == nil || timeouts.ReadHeader <= 0 || timeouts.Read <= 0 || timeouts.Write <= 0 || timeouts.Idle <= 0 {
		return nil, errors.New("handler and positive timeouts are required")
	}
	return &http.Server{
		Handler: handler,
		ReadHeaderTimeout: timeouts.ReadHeader,
		ReadTimeout: timeouts.Read,
		WriteTimeout: timeouts.Write,
		IdleTimeout: timeouts.Idle,
	}, nil
}

func Serve(ctx context.Context, server *http.Server, listener net.Listener, shutdownTimeout time.Duration) error {
	if ctx == nil || server == nil || listener == nil || shutdownTimeout <= 0 {
		return errors.New("context, server, listener and positive shutdown timeout are required")
	}
	serveErrors := make(chan error, 1)
	go func() { serveErrors <- server.Serve(listener) }()
	select {
	case err := <-serveErrors:
		return err
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		shutdownErr := server.Shutdown(shutdownContext)
		if shutdownErr != nil {
			_ = server.Close()
		}
		serveErr := <-serveErrors
		if shutdownErr != nil {
			return shutdownErr
		}
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			return serveErr
		}
		return nil
	}
}
`
}

func jobwatchHTTPTableTestSolution() string {
	return `package adminserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerContract(t *testing.T) {
	handler, err := NewHandler(Dependencies{
		Ready: func(context.Context) bool { return true },
		LookupJob: func(_ context.Context, id string) (string, bool) { return "queued", id == "job-1" },
		NextTraceID: func() string { return "test-trace" },
	})
	if err != nil { t.Fatal(err) }
	tests := []struct {
		name string
		method string
		path string
		want int
	}{
		{name: "ready", method: http.MethodGet, path: "/readyz", want: http.StatusNoContent},
		{name: "job", method: http.MethodGet, path: "/jobs/job-1", want: http.StatusOK},
		{name: "missing", method: http.MethodGet, path: "/jobs/missing", want: http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(testCase.method, testCase.path, nil))
			if recorder.Code != testCase.want { t.Fatalf("status = %d", recorder.Code) }
		})
	}
}
`
}
