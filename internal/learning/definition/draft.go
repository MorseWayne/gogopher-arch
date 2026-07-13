package definition

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

type draftDefinition[T any] struct {
	SourcePath string
	BundlePath string
	Raw        []byte
	Document   T
}

type draftSet struct {
	Capabilities []draftDefinition[capabilityDocument]
	Activities   []draftDefinition[activityDocument]
	Tasks        []draftDefinition[taskDocument]
}

func loadDraftSet(contentDir, activitySet string) (draftSet, error) {
	if !isCleanRelativePath(activitySet) || filepath.Base(activitySet) != activitySet {
		return draftSet{}, fmt.Errorf("activity set must be a clean path segment")
	}
	validator, err := NewValidator(os.DirFS(filepath.Join(contentDir, "schemas")))
	if err != nil {
		return draftSet{}, err
	}

	allCapabilities, err := loadCapabilities(filepath.Join(contentDir, "capabilities"), validator)
	if err != nil {
		return draftSet{}, err
	}
	activities, err := loadActivities(filepath.Join(contentDir, "activities", activitySet), activitySet, validator)
	if err != nil {
		return draftSet{}, err
	}
	allTasks, err := loadTasks(filepath.Join(contentDir, "tasks"), validator)
	if err != nil {
		return draftSet{}, err
	}

	capabilityByRef := make(map[string]draftDefinition[capabilityDocument], len(allCapabilities))
	for _, capability := range allCapabilities {
		key := referenceKey(capability.Document.ID, capability.Document.Version)
		if _, exists := capabilityByRef[key]; exists {
			return draftSet{}, fmt.Errorf("duplicate capability %s version %d", capability.Document.ID, capability.Document.Version)
		}
		capabilityByRef[key] = capability
	}
	taskByRef := make(map[string]draftDefinition[taskDocument], len(allTasks))
	for _, task := range allTasks {
		key := referenceKey(task.Document.ID, task.Document.Version)
		if _, exists := taskByRef[key]; exists {
			return draftSet{}, fmt.Errorf("duplicate task %s version %d", task.Document.ID, task.Document.Version)
		}
		taskByRef[key] = task
	}

	selectedCapabilities := make(map[string]draftDefinition[capabilityDocument])
	var includeCapability func(versionedRef) error
	includeCapability = func(ref versionedRef) error {
		key := referenceKey(ref.ID, ref.Version)
		if _, selected := selectedCapabilities[key]; selected {
			return nil
		}
		capability, exists := capabilityByRef[key]
		if !exists {
			return fmt.Errorf("capability reference %s version %d does not exist", ref.ID, ref.Version)
		}
		selectedCapabilities[key] = capability
		for _, prerequisite := range append(append([]versionedRef(nil), capability.Document.Prerequisites.Hard...), capability.Document.Prerequisites.Recommended...) {
			if err := includeCapability(prerequisite); err != nil {
				return fmt.Errorf("capability %s version %d prerequisite: %w", capability.Document.ID, capability.Document.Version, err)
			}
		}
		return nil
	}

	selectedTasks := make(map[string]draftDefinition[taskDocument])
	seenActivities := make(map[string]struct{}, len(activities))
	for _, activity := range activities {
		key := referenceKey(activity.Document.ID, activity.Document.Version)
		if _, exists := seenActivities[key]; exists {
			return draftSet{}, fmt.Errorf("duplicate activity %s version %d", activity.Document.ID, activity.Document.Version)
		}
		seenActivities[key] = struct{}{}
		for _, capabilityRef := range activity.Document.CapabilityRefs {
			if err := includeCapability(capabilityRef); err != nil {
				return draftSet{}, fmt.Errorf("activity %s version %d: %w", activity.Document.ID, activity.Document.Version, err)
			}
		}
		taskKey := referenceKey(activity.Document.TaskRef.ID, activity.Document.TaskRef.Version)
		task, exists := taskByRef[taskKey]
		if !exists {
			return draftSet{}, fmt.Errorf("activity %s version %d references missing task %s version %d", activity.Document.ID, activity.Document.Version, activity.Document.TaskRef.ID, activity.Document.TaskRef.Version)
		}
		if existing, selected := selectedTasks[taskKey]; selected && existing.SourcePath != task.SourcePath {
			return draftSet{}, fmt.Errorf("task %s version %d resolves to multiple sources", task.Document.ID, task.Document.Version)
		}
		selectedTasks[taskKey] = task
		if err := validateEvidenceRules(activity.Document, task.Document); err != nil {
			return draftSet{}, err
		}
	}

	capabilities := mapValues(selectedCapabilities)
	tasks := mapValues(selectedTasks)
	sortDrafts(capabilities)
	sortDrafts(activities)
	sortDrafts(tasks)
	if err := validateHardPrerequisites(capabilities); err != nil {
		return draftSet{}, err
	}
	return draftSet{Capabilities: capabilities, Activities: activities, Tasks: tasks}, nil
}

func loadCapabilities(root string, validator *Validator) ([]draftDefinition[capabilityDocument], error) {
	var definitions []draftDefinition[capabilityDocument]
	err := filepath.WalkDir(root, func(sourcePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("capability definitions may not contain symlink %q", sourcePath)
		}
		if entry.IsDir() || filepath.Ext(sourcePath) != ".json" {
			return nil
		}
		relative, err := filepath.Rel(root, sourcePath)
		if err != nil {
			return err
		}
		definition, err := readDraft[capabilityDocument](sourcePath, filepath.ToSlash(filepath.Join("capabilities", relative)), KindCapability, validator)
		if err != nil {
			return err
		}
		definitions = append(definitions, definition)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("load capabilities: %w", err)
	}
	return definitions, nil
}

func loadActivities(root, activitySet string, validator *Validator) ([]draftDefinition[activityDocument], error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read activity set %q: %w", activitySet, err)
	}
	definitions := make([]draftDefinition[activityDocument], 0, len(entries))
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("activity definitions may not contain symlink %q", filepath.Join(root, entry.Name()))
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		sourcePath := filepath.Join(root, entry.Name())
		definition, err := readDraft[activityDocument](sourcePath, filepath.ToSlash(filepath.Join("activities", activitySet, entry.Name())), KindActivity, validator)
		if err != nil {
			return nil, err
		}
		if entry.Name() != definition.Document.ID+".json" {
			return nil, fmt.Errorf("activity %s version %d must use file name %s.json", definition.Document.ID, definition.Document.Version, definition.Document.ID)
		}
		definitions = append(definitions, definition)
	}
	if len(definitions) == 0 {
		return nil, fmt.Errorf("activity set %q contains no definitions", activitySet)
	}
	return definitions, nil
}

func loadTasks(root string, validator *Validator) ([]draftDefinition[taskDocument], error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read tasks: %w", err)
	}
	definitions := make([]draftDefinition[taskDocument], 0, len(entries))
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("task definitions may not contain symlink %q", filepath.Join(root, entry.Name()))
		}
		if !entry.IsDir() {
			continue
		}
		taskDir := filepath.Join(root, entry.Name())
		sourcePath := filepath.Join(taskDir, "task.json")
		definition, err := readDraft[taskDocument](sourcePath, filepath.ToSlash(filepath.Join("tasks", entry.Name(), "task.json")), KindTask, validator)
		if err != nil {
			return nil, err
		}
		if entry.Name() != definition.Document.ID {
			return nil, fmt.Errorf("task %s version %d must use directory name %s", definition.Document.ID, definition.Document.Version, definition.Document.ID)
		}
		if err := ValidateTaskAssets(taskDir, definition.Raw); err != nil {
			return nil, fmt.Errorf("task %s version %d: %w", definition.Document.ID, definition.Document.Version, err)
		}
		definitions = append(definitions, definition)
	}
	return definitions, nil
}

func readDraft[T any](sourcePath, bundlePath string, kind Kind, validator *Validator) (draftDefinition[T], error) {
	raw, err := os.ReadFile(sourcePath)
	if err != nil {
		return draftDefinition[T]{}, fmt.Errorf("read %s: %w", sourcePath, err)
	}
	if err := validator.Validate(kind, raw); err != nil {
		return draftDefinition[T]{}, fmt.Errorf("%s: %w", sourcePath, err)
	}
	var document T
	if err := json.Unmarshal(raw, &document); err != nil {
		return draftDefinition[T]{}, fmt.Errorf("parse %s: %w", sourcePath, err)
	}
	return draftDefinition[T]{SourcePath: sourcePath, BundlePath: bundlePath, Raw: raw, Document: document}, nil
}

func validateEvidenceRules(activity activityDocument, task taskDocument) error {
	if len(activity.EvidenceRules) != len(task.AssessmentRules) {
		return fmt.Errorf("activity %s version %d has %d evidence rules but task %s version %d has %d assessment rules", activity.ID, activity.Version, len(activity.EvidenceRules), task.ID, task.Version, len(task.AssessmentRules))
	}
	assessmentRules := make(map[string]struct {
		capabilities map[string]struct{}
		evidenceType string
		condition    string
	}, len(task.AssessmentRules))
	for _, rule := range task.AssessmentRules {
		if _, exists := assessmentRules[rule.RuleID]; exists {
			return fmt.Errorf("task %s version %d has duplicate assessment rule %q", task.ID, task.Version, rule.RuleID)
		}
		capabilities := make(map[string]struct{}, len(rule.CapabilityRefs))
		for _, capability := range rule.CapabilityRefs {
			capabilities[referenceKey(capability.ID, capability.Version)] = struct{}{}
		}
		assessmentRules[rule.RuleID] = struct {
			capabilities map[string]struct{}
			evidenceType string
			condition    string
		}{capabilities: capabilities, evidenceType: rule.EvidenceType, condition: rule.Condition}
	}
	seen := make(map[string]struct{}, len(activity.EvidenceRules))
	for _, rule := range activity.EvidenceRules {
		if _, exists := seen[rule.RuleID]; exists {
			return fmt.Errorf("activity %s version %d has duplicate evidence rule %q", activity.ID, activity.Version, rule.RuleID)
		}
		seen[rule.RuleID] = struct{}{}
		assessment, exists := assessmentRules[rule.RuleID]
		if !exists {
			return fmt.Errorf("activity %s version %d evidence rule %q is missing from task %s version %d", activity.ID, activity.Version, rule.RuleID, task.ID, task.Version)
		}
		if _, exists := assessment.capabilities[referenceKey(rule.CapabilityID, rule.CapabilityVersion)]; !exists || assessment.evidenceType != rule.EvidenceType || assessment.condition != rule.Result {
			return fmt.Errorf("activity %s version %d evidence rule %q does not match task assessment rule", activity.ID, activity.Version, rule.RuleID)
		}
	}
	return nil
}

func validateHardPrerequisites(capabilities []draftDefinition[capabilityDocument]) error {
	byRef := make(map[string]capabilityDocument, len(capabilities))
	for _, capability := range capabilities {
		byRef[referenceKey(capability.Document.ID, capability.Document.Version)] = capability.Document
	}
	for _, capability := range capabilities {
		prerequisites := append(append([]versionedRef(nil), capability.Document.Prerequisites.Hard...), capability.Document.Prerequisites.Recommended...)
		for _, prerequisite := range prerequisites {
			if _, exists := byRef[referenceKey(prerequisite.ID, prerequisite.Version)]; !exists {
				return fmt.Errorf("capability %s version %d prerequisite %s version %d is missing", capability.Document.ID, capability.Document.Version, prerequisite.ID, prerequisite.Version)
			}
		}
	}
	state := make(map[string]uint8, len(capabilities))
	var visit func(string) error
	visit = func(key string) error {
		switch state[key] {
		case 1:
			capability := byRef[key]
			return fmt.Errorf("hard prerequisite cycle includes capability %s version %d", capability.ID, capability.Version)
		case 2:
			return nil
		}
		state[key] = 1
		capability := byRef[key]
		for _, prerequisite := range capability.Prerequisites.Hard {
			prerequisiteKey := referenceKey(prerequisite.ID, prerequisite.Version)
			if _, exists := byRef[prerequisiteKey]; !exists {
				return fmt.Errorf("capability %s version %d hard prerequisite %s version %d is missing", capability.ID, capability.Version, prerequisite.ID, prerequisite.Version)
			}
			if err := visit(prerequisiteKey); err != nil {
				return err
			}
		}
		state[key] = 2
		return nil
	}
	for key := range byRef {
		if err := visit(key); err != nil {
			return err
		}
	}
	return nil
}

func sortDrafts[T any](definitions []draftDefinition[T]) {
	sort.Slice(definitions, func(i, j int) bool {
		return definitions[i].BundlePath < definitions[j].BundlePath
	})
}

func mapValues[T any](values map[string]draftDefinition[T]) []draftDefinition[T] {
	result := make([]draftDefinition[T], 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}
