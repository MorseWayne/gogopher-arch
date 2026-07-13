package evaluation

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"strconv"
	"strings"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

type Generator struct {
	registry *definition.Registry
}

func NewGenerator(registry *definition.Registry) (*Generator, error) {
	if registry == nil {
		return nil, fmt.Errorf("definition registry is required")
	}
	return &Generator{registry: registry}, nil
}

func (g *Generator) Generate(frozen submission.Submission, terminal execution.Execution) ([]execution.RuleResult, error) {
	if terminal.SubmissionID != frozen.ID || terminal.AttemptID != frozen.AttemptID ||
		terminal.TaskID != frozen.TaskID || terminal.TaskVersion != frozen.TaskVersion ||
		terminal.TaskHash != frozen.TaskHash {
		return nil, fmt.Errorf("terminal execution does not match frozen submission")
	}
	if terminal.Status == execution.ExecutionInfraFailed || terminal.Response == nil {
		return nil, fmt.Errorf("infrastructure-failed or incomplete execution cannot produce rule results")
	}
	if terminal.Status != terminal.Response.Status {
		return nil, fmt.Errorf("execution status does not match its terminal response")
	}
	if err := terminal.Response.Validate(); err != nil {
		return nil, fmt.Errorf("validate terminal response: %w", err)
	}
	task, err := g.registry.ExecutionTask(frozen.ReleaseID, frozen.TaskID, frozen.TaskVersion)
	if err != nil {
		return nil, fmt.Errorf("resolve frozen assessment rules: %w", err)
	}
	if task.BundleHash != frozen.TaskHash {
		return nil, fmt.Errorf("frozen task bundle hash mismatch")
	}
	stages := make(map[execution.Stage]execution.StageResult, len(terminal.Response.Stages))
	for _, stage := range terminal.Response.Stages {
		stages[stage.Stage] = stage
	}
	results := make([]execution.RuleResult, 0, len(task.AssessmentRules))
	for _, rule := range task.AssessmentRules {
		result, err := evaluateRule(rule, task, frozen.Workspace, stages, terminal.ID)
		if err != nil {
			return nil, fmt.Errorf("evaluate rule %q: %w", rule.RuleID, err)
		}
		if err := result.Validate(); err != nil {
			return nil, fmt.Errorf("validate rule %q result: %w", rule.RuleID, err)
		}
		results = append(results, result)
	}
	return results, nil
}

func evaluateRule(rule definition.AssessmentRule, task definition.ExecutionTask, workspace map[string]string, stages map[execution.Stage]execution.StageResult, executionID string) (execution.RuleResult, error) {
	if rule.Condition != "passed" {
		return execution.RuleResult{}, fmt.Errorf("unsupported condition %q", rule.Condition)
	}
	stage := execution.Stage(rule.Stage)
	result := execution.RuleResult{
		RuleID: rule.RuleID, Stage: stage, Package: rule.Selector.Package,
		Test: rule.Selector.Test, ExecutionID: executionID,
	}
	var status execution.RuleStatus
	var analyzer string
	if stage == execution.StageAST {
		status, analyzer = evaluateASTRule(rule, task, workspace, stages)
	} else {
		status = evaluateExecutionRule(rule, stages)
	}
	result.Status = status
	result.Analyzer = analyzer
	result.Summary = ruleSummary(rule.RuleID, status)
	return result, nil
}

func evaluateExecutionRule(rule definition.AssessmentRule, stages map[execution.Stage]execution.StageResult) execution.RuleStatus {
	stage, exists := stages[execution.Stage(rule.Stage)]
	if !exists {
		return execution.RuleNotEvaluated
	}
	if rule.Selector.Test != "" {
		if stage.TimedOut || stage.OutputTruncated {
			return execution.RuleNotEvaluated
		}
		for index := len(stage.TestEvents) - 1; index >= 0; index-- {
			event := stage.TestEvents[index]
			if event.Test != rule.Selector.Test {
				continue
			}
			if event.Action == "pass" {
				return execution.RulePassed
			}
			if event.Action == "fail" || event.Action == "skip" {
				return execution.RuleFailed
			}
		}
		return execution.RuleNotEvaluated
	}
	if stage.TimedOut || stage.OutputTruncated {
		return execution.RuleFailed
	}
	if rule.Selector.ExitCode != nil && stage.ExitCode != *rule.Selector.ExitCode {
		return execution.RuleFailed
	}
	if stage.Status == execution.StagePassed {
		return execution.RulePassed
	}
	return execution.RuleFailed
}

func evaluateASTRule(rule definition.AssessmentRule, task definition.ExecutionTask, workspace map[string]string, stages map[execution.Stage]execution.StageResult) (execution.RuleStatus, string) {
	if rule.Selector.DeferredCall != "" {
		build, exists := stages[execution.StageBuild]
		if !exists || build.Status != execution.StagePassed {
			return execution.RuleNotEvaluated, "go_ast_deferred_call"
		}
		if hasDeferredCall(workspace[rule.Selector.File], rule.Selector.DeferredCall) {
			return execution.RulePassed, "go_ast_deferred_call"
		}
		return execution.RuleFailed, "go_ast_deferred_call"
	}
	if rule.Selector.MinimumCases > 0 {
		visible, exists := stages[execution.StageVisibleTest]
		if !exists {
			return execution.RuleNotEvaluated, "go_ast_table_tests"
		}
		if visible.Status != execution.StagePassed {
			return execution.RuleFailed, "go_ast_table_tests"
		}
		if hasLearnerTableTests(task, workspace, rule.Selector) {
			return execution.RulePassed, "go_ast_table_tests"
		}
		return execution.RuleFailed, "go_ast_table_tests"
	}
	return execution.RuleFailed, "go_ast_unknown"
}

func hasDeferredCall(source, callName string) bool {
	file, err := parser.ParseFile(token.NewFileSet(), "source.go", source, 0)
	if err != nil {
		return false
	}
	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		deferStatement, ok := node.(*ast.DeferStmt)
		if !ok {
			return true
		}
		selector, ok := deferStatement.Call.Fun.(*ast.SelectorExpr)
		if ok && selector.Sel.Name == callName {
			found = true
		}
		return !found
	})
	return found
}

func hasLearnerTableTests(task definition.ExecutionTask, workspace map[string]string, selector definition.AssessmentSelector) bool {
	baselines := make(map[string]string)
	for _, asset := range task.Files {
		if asset.Editable {
			baselines[asset.Path] = asset.Content
		}
	}
	for filePath, baseline := range baselines {
		if !strings.HasSuffix(filePath, "_test.go") || !selectorMatchesFile(selector, filePath) {
			continue
		}
		source, exists := workspace[filePath]
		if !exists || source == baseline {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), filePath, source, 0)
		if err != nil || !containsRunCall(file) || maximumNamedTableCases(file) < selector.MinimumCases {
			continue
		}
		return true
	}
	return false
}

func selectorMatchesFile(selector definition.AssessmentSelector, filePath string) bool {
	if selector.File != "" {
		return selector.File == filePath
	}
	if selector.Glob == "" {
		return false
	}
	if matched, _ := path.Match(selector.Glob, filePath); matched {
		return true
	}
	if strings.HasPrefix(selector.Glob, "**/") {
		matched, _ := path.Match(strings.TrimPrefix(selector.Glob, "**/"), path.Base(filePath))
		return matched
	}
	return false
}

func containsRunCall(file *ast.File) bool {
	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if ok && selector.Sel.Name == "Run" {
			found = true
		}
		return !found
	})
	return found
}

func maximumNamedTableCases(file *ast.File) int {
	maximum := 0
	ast.Inspect(file, func(node ast.Node) bool {
		literal, ok := node.(*ast.CompositeLit)
		if !ok {
			return true
		}
		if _, ok := literal.Type.(*ast.ArrayType); !ok {
			return true
		}
		named := 0
		for _, element := range literal.Elts {
			row, ok := element.(*ast.CompositeLit)
			if !ok || !hasCaseName(row) {
				continue
			}
			named++
		}
		if named > maximum {
			maximum = named
		}
		return true
	})
	return maximum
}

func hasCaseName(row *ast.CompositeLit) bool {
	for index, element := range row.Elts {
		if keyed, ok := element.(*ast.KeyValueExpr); ok {
			identifier, ok := keyed.Key.(*ast.Ident)
			if ok && identifier.Name == "name" && nonemptyString(keyed.Value) {
				return true
			}
			continue
		}
		if index == 0 && nonemptyString(element) {
			return true
		}
	}
	return false
}

func nonemptyString(expression ast.Expr) bool {
	literal, ok := expression.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return false
	}
	value, err := strconv.Unquote(literal.Value)
	return err == nil && value != ""
}

func ruleSummary(ruleID string, status execution.RuleStatus) string {
	switch status {
	case execution.RulePassed:
		return fmt.Sprintf("rule %s passed", ruleID)
	case execution.RuleFailed:
		return fmt.Sprintf("rule %s failed", ruleID)
	default:
		return fmt.Sprintf("rule %s was not evaluated", ruleID)
	}
}
