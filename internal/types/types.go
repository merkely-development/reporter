package types

type PREvidence struct {
	MergeCommit string   `json:"merge_commit"`
	URL         string   `json:"url"`
	State       string   `json:"state"`
	Approvers   []string `json:"approvers"`
	Author      string   `json:"author"`
	// Compliant   *bool    `json:"is_compliant,omitempty"`
	// ReasonsForNonCompliance []string `json:"reasons_for_non_compliance,omitempty"`
	// LastCommit             string `json:"lastCommit"`
	// LastCommitter          string `json:"lastCommitter"`
	// SelfApproved           bool   `json:"selfApproved"`
}

type PRRetriever interface {
	PREvidenceForCommit(string) ([]*PREvidence, error)
}

// type PRAuthor struct {
// 	Login string `json:"login"`
// }
