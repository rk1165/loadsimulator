package config

import (
	"fmt"

	"github.com/rk1165/loadsimulator/internal/assets"
	"github.com/rk1165/loadsimulator/internal/logger"
	"github.com/rk1165/loadsimulator/internal/types"
	"gopkg.in/yaml.v3"
)

func LoadScenarios[T any](fileName string) (map[string]T, error) {
	logger.InfoLog.Printf("Loading Scenarios from file : %s", fileName)
	scenarios := make(map[string]T)
	b, err := assets.FS.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenarios file=%s error=[%v]", fileName, err)
	}

	if err := yaml.Unmarshal(b, &scenarios); err != nil {
		return nil, fmt.Errorf("failed to unmarshal scenarios file=%s error=[%v]", fileName, err)
	}
	logger.InfoLog.Printf("Loaded all scenarios successfully from %s", fileName)
	return scenarios, nil
}

func LoadTestConfig[T types.Provider](configFile string, scenarioName string) (T, types.Config, error) {
	logger.InfoLog.Printf("Loading TestConfig scenario=%s", scenarioName)
	var zero T

	scenarios, err := LoadScenarios[T](configFile)
	if err != nil {
		return zero, types.Config{}, err
	}
	scenario, ok := scenarios[scenarioName]
	if !ok {
		return zero, types.Config{}, fmt.Errorf("scenario %s not found in %s", scenarioName, configFile)
	}
	cfg := types.Config{
		Name:        scenarioName,
		RatePerSec:  scenario.GetRatePerSecond(),
		Duration:    scenario.GetDuration(),
		Concurrency: scenario.GetConcurrentRequests(),
		InfoLog:     logger.InfoLog,
		ErrorLog:    logger.ErrorLog,
		//Jitter:      scenario.GetJitter(),
	}
	logger.InfoLog.Printf("Loaded TestConfig scenario=%s", scenarioName)
	return scenario, cfg, nil
}
