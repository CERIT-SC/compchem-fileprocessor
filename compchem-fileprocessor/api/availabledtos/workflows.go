package availabledtos

type AvailableWorkflowsRequest struct {
	Files []KeyAndType `json:"files"`
}

type KeyAndType struct {
	FileKey  string `json:"key"`
	Mimetype string `json:"mimetype"`
}

type AvailableWorkflowsResponse struct {
	Workflows []AvailableWorkflow `json:"workflows"`
}

type AvailableWorkflow struct {
	Name     string   `json:"name"`
	Mimetype string   `json:"mimetype"`
	Files    []string `json:"files"`
}
