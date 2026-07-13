package sandbox

import (
	"net/http"

	sandboxrunner "github.com/MorseWayne/gogopher-arch/internal/sandbox"
)

func Build() http.Handler {
	return sandboxrunner.NewHTTPHandler(sandboxrunner.NewRunner(sandboxrunner.RunnerOptions{}))
}
