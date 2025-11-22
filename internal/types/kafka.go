package types

type KafkaConfig struct {
	BaseConfig     `yaml:",inline"`
	ClientId       string `yaml:"clientId"`
	ClientSecret   string `yaml:"clientSecret"`
	UserName       string `yaml:"username"`
	Password       string `yaml:"password"`
	Authentication string `yaml:"authentication"`
	Topic          string `yaml:"topic"`
	Broker         string `yaml:"broker"`
}

type KafkaScenarios map[string]KafkaConfig
