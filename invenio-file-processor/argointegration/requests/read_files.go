package requests

const ReadFilesTemplate = "read-files-%s-%s"

func NewReadFilesWorkflow(name string,
	deps []string,
	predecessor string,
	recordId string,
	workflowId string,
) *Task {
	return &Task{}
}
