package rest

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rk1165/loadsimulator/internal/load"
	"github.com/rk1165/loadsimulator/internal/logger"
	"github.com/rk1165/loadsimulator/internal/types"
)

type LoadPostApi struct {
	load.BaseLoad
	url                string
	headers            map[string]string
	body               string
	expectedStatusCode int
	log                load.Log
	replaceParams      []types.KV
}

func NewPost(apiConfig types.ApiConfig, cfg types.Config) *LoadPostApi {
	postApiLoad := &LoadPostApi{
		url:                apiConfig.ResolveEndPoint(),
		expectedStatusCode: apiConfig.ExpectedStatusCode,
		headers:            resolveHeaders(apiConfig),
		BaseLoad:           load.NewBaseLoad(cfg),
		log:                logger.CreateLoadLog(cfg.Name),
		replaceParams:      apiConfig.ReplaceParams,
	}

	if apiConfig.Method == "POST" {
		postApiLoad.body = apiConfig.ResolveBody()
	}
	return postApiLoad
}

func (p *LoadPostApi) Execute(ctx context.Context, id uint64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// replace fields in the body
	newBody := p.body
	if p.replaceParams != nil {
		for _, v := range p.replaceParams {
			if v.Value == "UUID" {
				newBody = strings.Replace(newBody, v.Key, uuid.New().String(), -1)
			}
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, strings.NewReader(newBody))

	if err != nil {
		return err
	}

	for k, v := range p.headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	duration := time.Since(start)
	if p.Success(resp) {
		p.Record(duration, true)
		p.log.InfoLog.Printf("[HTTP POST] id=%d status=%d elapsed=%s", id, resp.StatusCode, duration)
	} else {
		p.Record(duration, false)
		p.log.ErrorLog.Printf("[HTTP POST] id=%d status=%d elapsed=%s", id, resp.StatusCode, duration)
	}
	return nil
}

func (p *LoadPostApi) Success(response any) bool {
	apiResponse := response.(*http.Response)
	return apiResponse.StatusCode == p.expectedStatusCode
}

func (p *LoadPostApi) CalculateStats() *load.Stats {
	return p.BaseLoad.CalculateStats()
}
