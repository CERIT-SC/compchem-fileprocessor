package requests

type Task struct {
	Name              string            `yaml:"name"`
	Dependencies      string          `yaml:"dependencies"`
	TemplateReference TemplateReference `yaml:"templateRef"`
	Arguments         ParametersAndArtifacts         `yaml:"arguments"`
}

type Parameter struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Artifact struct {
	Name string `json:"name"`
	From string `json:"from"`
}

type ParametersAndArtifacts struct {
	Parameters []Parameter `yaml:"parameters"`
	Artifacts  []Artifact  `yaml:"artifacts"`
}

type TemplateReference struct {
	Name     string `yaml:"name"`
	Template string `yaml:"template"`
}
