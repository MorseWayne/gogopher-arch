package definition

import "strconv"

const ReleaseManifestSchemaVersion = 1

type ReleaseManifest struct {
	ReleaseID     string               `json:"release_id"`
	CreatedAt     string               `json:"created_at"`
	SchemaVersion int                  `json:"schema_version"`
	ActivitySet   string               `json:"activity_set"`
	Definitions   []ManifestDefinition `json:"definitions"`
	Assets        []ManifestAsset      `json:"assets"`
	BundleHash    string               `json:"bundle_hash"`
}

type ManifestDefinition struct {
	Kind        Kind   `json:"kind"`
	ID          string `json:"id"`
	Version     int    `json:"version"`
	Path        string `json:"path"`
	ContentHash string `json:"content_hash"`
	BundleHash  string `json:"bundle_hash,omitempty"`
	RuleSetHash string `json:"rule_set_hash,omitempty"`
}

type ManifestAsset struct {
	TaskID        string `json:"task_id"`
	TaskVersion   int    `json:"task_version"`
	Source        string `json:"source"`
	WorkspacePath string `json:"workspace_path"`
	BundlePath    string `json:"bundle_path"`
	Role          string `json:"role"`
	SHA256        string `json:"sha256"`
}

type versionedRef struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
}

type capabilityDocument struct {
	ID            string `json:"id"`
	Version       int    `json:"version"`
	Prerequisites struct {
		Hard        []versionedRef `json:"hard"`
		Recommended []versionedRef `json:"recommended"`
	} `json:"prerequisites"`
}

type activityDocument struct {
	ID             string         `json:"id"`
	Version        int            `json:"version"`
	CapabilityRefs []versionedRef `json:"capability_refs"`
	TaskRef        versionedRef   `json:"task_ref"`
	EvidenceRules  []struct {
		RuleID            string `json:"rule_id"`
		CapabilityID      string `json:"capability_id"`
		CapabilityVersion int    `json:"capability_version"`
		EvidenceType      string `json:"evidence_type"`
		Result            string `json:"result"`
	} `json:"evidence_rules"`
}

type taskDocument struct {
	ID              string     `json:"id"`
	Version         int        `json:"version"`
	Files           []taskFile `json:"files"`
	AssessmentRules []struct {
		RuleID         string         `json:"rule_id"`
		CapabilityRefs []versionedRef `json:"capability_refs"`
		EvidenceType   string         `json:"evidence_type"`
		Condition      string         `json:"condition"`
	} `json:"assessment_rules"`
}

func definitionKey(kind Kind, id string, version int) string {
	return string(kind) + "\x00" + id + "\x00" + strconv.Itoa(version)
}

func referenceKey(id string, version int) string {
	return id + "\x00" + strconv.Itoa(version)
}
