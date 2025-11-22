package types

type S3Config struct {
	BaseConfig `yaml:",inline"`
	Bucket     string `yaml:"bucket"`
	Key        string `yaml:"key"`
	Extension  string `yaml:"extension"`
	Region     string `yaml:"region"`
}

type S3Scenarios map[string]S3Config
