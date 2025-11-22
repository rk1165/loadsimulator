package types

type MessageAttribute struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
	Type  string `yaml:"type"` // "String", "Number", "Binary"
}

type SqsConfig struct {
	BaseConfig        `yaml:",inline"`
	Queue             string             `yaml:"queue"`
	Region            string             `yaml:"region"`
	MessageAttributes []MessageAttribute `yaml:"messageAttributes"`
}

type SqsScenarios map[string]SqsConfig
