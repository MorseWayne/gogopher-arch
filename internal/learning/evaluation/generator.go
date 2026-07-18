package evaluation

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"strconv"
	"strings"
	"unicode/utf8"

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
		result, err := evaluateRule(rule, task, frozen.Workspace, frozen.Explanation, stages, terminal.ID)
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

func evaluateRule(rule definition.AssessmentRule, task definition.ExecutionTask, workspace map[string]string, explanation string, stages map[execution.Stage]execution.StageResult, executionID string) (execution.RuleResult, error) {
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
	if stage == execution.StageExplanation {
		if rule.Selector.MinimumChars < 1 {
			return execution.RuleResult{}, fmt.Errorf("explanation selector requires minimum_chars")
		}
		if explanationSatisfies(explanation, rule.Selector) {
			status = execution.RulePassed
		} else {
			status = execution.RuleFailed
		}
	} else if stage == execution.StageAST {
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
	if len(rule.Selector.RequiredCalls) > 0 {
		build, exists := stages[execution.StageBuild]
		if !exists || build.Status != execution.StagePassed {
			return execution.RuleNotEvaluated, "go_ast_required_calls"
		}
		if hasRequiredCalls(workspace[rule.Selector.File], rule.Selector.RequiredCalls) {
			return execution.RulePassed, "go_ast_required_calls"
		}
		return execution.RuleFailed, "go_ast_required_calls"
	}
	if len(rule.Selector.RequiredFiles) > 0 {
		build, exists := stages[execution.StageBuild]
		if !exists || build.Status != execution.StagePassed {
			return execution.RuleNotEvaluated, "workspace_required_files"
		}
		for _, filePath := range rule.Selector.RequiredFiles {
			if strings.TrimSpace(workspace[filePath]) == "" {
				return execution.RuleFailed, "workspace_required_files"
			}
		}
		return execution.RulePassed, "workspace_required_files"
	}
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
	if rule.Selector.DocumentedExports {
		build, exists := stages[execution.StageBuild]
		if !exists || build.Status != execution.StagePassed {
			return execution.RuleNotEvaluated, "go_ast_documented_exports"
		}
		if hasDocumentedExports(workspace[rule.Selector.File]) {
			return execution.RulePassed, "go_ast_documented_exports"
		}
		return execution.RuleFailed, "go_ast_documented_exports"
	}
	if rule.Selector.Interface != "" {
		build, exists := stages[execution.StageBuild]
		if !exists || build.Status != execution.StagePassed {
			return execution.RuleNotEvaluated, "go_ast_minimal_interface"
		}
		if hasMinimalInterface(workspace[rule.Selector.File], rule.Selector.Interface, rule.Selector.MaximumMethods) {
			return execution.RulePassed, "go_ast_minimal_interface"
		}
		return execution.RuleFailed, "go_ast_minimal_interface"
	}
	if rule.Selector.GenericFunction != "" {
		build, exists := stages[execution.StageBuild]
		if !exists || build.Status != execution.StagePassed {
			return execution.RuleNotEvaluated, "go_ast_generic_function"
		}
		if hasGenericFunction(workspace[rule.Selector.File], rule.Selector.GenericFunction) {
			return execution.RulePassed, "go_ast_generic_function"
		}
		return execution.RuleFailed, "go_ast_generic_function"
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

func hasRequiredCalls(source string, required []string) bool {
	if len(required) == 0 {
		return false
	}
	file, err := parser.ParseFile(token.NewFileSet(), "source.go", source, 0)
	if err != nil {
		return false
	}
	found := make(map[string]struct{}, len(required))
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		owner, ok := selector.X.(*ast.Ident)
		if ok {
			found[owner.Name+"."+selector.Sel.Name] = struct{}{}
		}
		return true
	})
	for _, name := range required {
		if _, exists := found[name]; !exists {
			return false
		}
	}
	return true
}

func explanationSatisfies(explanation string, selector definition.AssessmentSelector) bool {
	trimmed := strings.TrimSpace(explanation)
	if utf8.RuneCountInString(trimmed) < selector.MinimumChars {
		return false
	}
	lower := strings.ToLower(trimmed)
	for _, term := range selector.RequiredTerms {
		term = strings.ToLower(strings.TrimSpace(term))
		if term == "" || !strings.Contains(lower, term) {
			return false
		}
	}
	return true
}

func hasMinimalInterface(source, name string, maximumMethods int) bool {
	if maximumMethods < 1 {
		return false
	}
	file, err := parser.ParseFile(token.NewFileSet(), "source.go", source, 0)
	if err != nil {
		return false
	}
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.TYPE {
			continue
		}
		for _, specification := range general.Specs {
			typeSpec, ok := specification.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != name {
				continue
			}
			interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				return false
			}
			methods := 0
			for _, field := range interfaceType.Methods.List {
				if len(field.Names) == 0 {
					methods++
				} else {
					methods += len(field.Names)
				}
			}
			return methods >= 1 && methods <= maximumMethods
		}
	}
	return false
}

func hasGenericFunction(source, name string) bool {
	file, err := parser.ParseFile(token.NewFileSet(), "source.go", source, 0)
	if err != nil {
		return false
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name.Name == name && function.Type.TypeParams != nil && len(function.Type.TypeParams.List) > 0 {
			return true
		}
	}
	return false
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

func hasDocumentedExports(source string) bool {
	file, err := parser.ParseFile(token.NewFileSet(), "source.go", source, parser.ParseComments)
	if err != nil || file.Doc == nil || !commentStartsWith(file.Doc, "Package "+file.Name.Name) {
		return false
	}
	found := false
	for _, declaration := range file.Decls {
		switch value := declaration.(type) {
		case *ast.FuncDecl:
			if !value.Name.IsExported() {
				continue
			}
			found = true
			if !commentStartsWith(value.Doc, value.Name.Name) {
				return false
			}
		case *ast.GenDecl:
			for _, specification := range value.Specs {
				var names []*ast.Ident
				var documentation *ast.CommentGroup
				switch item := specification.(type) {
				case *ast.TypeSpec:
					names = []*ast.Ident{item.Name}
					documentation = item.Doc
				case *ast.ValueSpec:
					names = item.Names
					documentation = item.Doc
				}
				if documentation == nil {
					documentation = value.Doc
				}
				for _, name := range names {
					if !name.IsExported() {
						continue
					}
					found = true
					if !commentStartsWith(documentation, name.Name) {
						return false
					}
				}
			}
		}
	}
	return found
}

func commentStartsWith(group *ast.CommentGroup, prefix string) bool {
	if group == nil {
		return false
	}
	return strings.HasPrefix(strings.TrimSpace(group.Text()), prefix)
}

func hasLearnerTableTests(task definition.ExecutionTask, workspace map[string]string, selector definition.AssessmentSelector) bool {
	baselines := make(map[string]string)
	declared := make(map[string]struct{}, len(task.Files))
	for _, asset := range task.Files {
		declared[asset.Path] = struct{}{}
		if asset.Editable {
			baselines[asset.Path] = asset.Content
		}
	}
	if task.WorkspacePolicy.AllowNewFiles {
		for filePath := range workspace {
			if _, exists := declared[filePath]; !exists {
				baselines[filePath] = ""
			}
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
