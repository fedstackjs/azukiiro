package common

import "encoding/json"

// Solution configuration
type ProblemConfigSolution struct {
	MaxSize *int `json:"maxSize,omitempty"`
}

// For submit form editor
type ProblemConfigSubmitFormEditor struct {
	Language string `json:"language"`
}

// For submit form metadata item type
type ProblemConfigSubmitFormMetadataItemType struct {
	Text   *struct{} `json:"text,omitempty"`
	Select *struct {
		Options []string `json:"options"`
	} `json:"select,omitempty"`
}

// For submit form metadata item
type ProblemConfigSubmitFormMetadataItem struct {
	Key         string                                  `json:"key"`
	Label       *string                                 `json:"label,omitempty"`
	Description *string                                 `json:"description,omitempty"`
	Type        ProblemConfigSubmitFormMetadataItemType `json:"type"`
}

// For submit form metadata
type ProblemConfigSubmitFormMetadata struct {
	Items []ProblemConfigSubmitFormMetadataItem `json:"items"`
}

// For submit form file type
type ProblemConfigSubmitFormFileType struct {
	Editor   *ProblemConfigSubmitFormEditor   `json:"editor,omitempty"`
	Metadata *ProblemConfigSubmitFormMetadata `json:"metadata,omitempty"`
}

// For submit form file
type ProblemConfigSubmitFormFile struct {
	Path        string                          `json:"path"`
	Label       *string                         `json:"label,omitempty"`
	Description *string                         `json:"description,omitempty"`
	Default     *string                         `json:"default,omitempty"`
	Type        ProblemConfigSubmitFormFileType `json:"type"`
}

// For submit form
type ProblemConfigSubmitForm struct {
	Files []ProblemConfigSubmitFormFile `json:"files"`
}

// For submit config
type ProblemConfigSubmit struct {
	Upload    *bool                    `json:"upload,omitempty"`
	ZipFolder *bool                    `json:"zipFolder,omitempty"`
	Form      *ProblemConfigSubmitForm `json:"form,omitempty"`
}

// For instance config
type ProblemConfigInstance struct {
	Adapter string          `json:"adapter"`
	Config  json.RawMessage `json:"config"`
}

// For judge config
type ProblemConfigJudge struct {
	Adapter string          `json:"adapter"`
	Config  json.RawMessage `json:"config"`
}

// Problem configuration
type ProblemConfig struct {
	Label         string                 `json:"label"`
	Solution      *ProblemConfigSolution `json:"solution,omitempty"`
	Judge         ProblemConfigJudge     `json:"judge"`
	Submit        *ProblemConfigSubmit   `json:"submit,omitempty"`
	InstanceLabel *string                `json:"instanceLabel,omitempty"`
	Instance      *ProblemConfigInstance `json:"instance,omitempty"`
	Variables     map[string]string      `json:"variables,omitempty"`
}
