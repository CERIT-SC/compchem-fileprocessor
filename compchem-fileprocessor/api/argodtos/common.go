package argodtos

type Task struct {
	Name              string                 `json:"name"`
	Dependencies      []string               `json:"dependencies"`
	TemplateReference TemplateReference      `json:"templateRef"`
	Arguments         ParametersAndArtifacts `json:"arguments"`
}

type Parameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Artifact struct {
	Name string `json:"name"`
	From string `json:"from"`
}

type ParametersAndArtifacts struct {
	Parameters []Parameter `json:"parameters"`
	Artifacts  []Artifact  `json:"artifacts"`
}

type TemplateReference struct {
	Name     string `json:"name"`
	Template string `json:"template"`
}
