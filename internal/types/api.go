package types

import (
	"fmt"
	"log"
	"strings"
)

type KV struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type ApiConfig struct {
	BaseConfig         `yaml:",inline"`
	Method             string `yaml:"method"`
	ClientId           string `yaml:"clientId"`
	ClientSecret       string `yaml:"clientSecret"`
	Scope              string `yaml:"scope"`
	BaseUrl            string `yaml:"baseUrl"`
	Endpoint           string `yaml:"endpoint"`
	ContentType        string `yaml:"contentType"`
	ExpectedStatusCode int    `yaml:"expectedStatusCode"`

	PathVariables []KV `yaml:"pathVariables"`
	QueryParams   []KV `yaml:"queryParams"`
	ReplaceParams []KV `yaml:"replaceParams"`
}

type ApiScenarios map[string]ApiConfig

func (a ApiConfig) ResolveEndPoint() string {
	url := a.BaseUrl + a.Endpoint
	if a.PathVariables != nil {
		for _, v := range a.PathVariables {
			placeholder := "{" + v.Key + "}"
			url = strings.Replace(url, placeholder, v.Value, -1)
		}
	}
	if a.QueryParams != nil {
		url = url + "?"
		var queryParams []string
		for _, v := range a.QueryParams {
			placeholder := fmt.Sprintf("%s=%s", v.Key, v.Value)
			queryParams = append(queryParams, placeholder)
		}
		url = url + strings.Join(queryParams, "&")
	}
	log.Printf("finalUrl=%s", url)
	return url
}
