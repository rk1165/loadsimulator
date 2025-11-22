package rest

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/rk1165/loadsimulator/internal/load"
	"github.com/rk1165/loadsimulator/internal/logger"
	"github.com/rk1165/loadsimulator/internal/types"
)

func newHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

var client = newHTTPClient(time.Duration(5000) * time.Millisecond)

type LoadGetApi struct {
	load.BaseLoad
	url                string
	headers            map[string]string
	expectedStatusCode int
	log                load.Log
}

func NewGet(apiConfig types.ApiConfig, cfg types.Config) *LoadGetApi {
	getApiLoad := &LoadGetApi{
		url:                apiConfig.ResolveEndPoint(),
		headers:            resolveHeaders(apiConfig),
		expectedStatusCode: apiConfig.ExpectedStatusCode,
		BaseLoad:           load.NewBaseLoad(cfg),
		log:                logger.CreateLoadLog(cfg.Name),
	}
	return getApiLoad
}

func (g *LoadGetApi) Execute(ctx context.Context, id uint64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.url, nil)

	if err != nil {
		return err
	}

	for k, v := range g.headers {
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
	if g.Success(resp) {
		g.Record(duration, true)
		g.log.InfoLog.Printf("[HTTP GET] requestId=%d status=%d elapsed=%s", id, resp.StatusCode, duration)
	} else {
		g.Record(duration, false)
		g.log.ErrorLog.Printf("[HTTP GET] requestId=%d status=%d elapsed=%s", id, resp.StatusCode, duration)
	}
	return nil
}

func (g *LoadGetApi) Success(response any) bool {
	apiResponse := response.(*http.Response)
	return apiResponse.StatusCode == g.expectedStatusCode
}

func (g *LoadGetApi) CalculateStats() *load.Stats {
	return g.BaseLoad.CalculateStats()
}
