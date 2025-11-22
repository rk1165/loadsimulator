package types

import (
	"log"
	"time"

	"github.com/rk1165/loadsimulator/internal/assets"
)

type Config struct {
	Name        string
	RatePerSec  int
	Duration    int
	Concurrency int
	Jitter      time.Duration
	InfoLog     *log.Logger
	ErrorLog    *log.Logger
}

type BaseConfig struct {
	FileName      string `yaml:"fileName"`
	RatePerSecond int    `yaml:"ratePerSec"`         // target operations per second
	Duration      int    `yaml:"duration"`           // Total time for the operations to run
	Concurrency   int    `yaml:"concurrentRequests"` // max in-flight tasks executing at the same time - implemented by starting that many worker goroutines generating the scheduled loads
	//Jitter        time.Duration							// max random +/- scheduled fire (optional)
}

type Provider interface {
	GetRatePerSecond() int
	GetConcurrentRequests() int
	GetDuration() int
	//GetJitter() time.Duration
}

func (b BaseConfig) GetRatePerSecond() int {
	return b.RatePerSecond
}

func (b BaseConfig) GetDuration() int {
	return b.Duration
}

func (b BaseConfig) GetConcurrentRequests() int {
	return b.Concurrency
}

func (b BaseConfig) ResolveBody() string {
	if len(b.FileName) == 0 {
		return ""
	}
	bytes, err := assets.FS.ReadFile(b.FileName)
	if err != nil {
		log.Fatalf("failed to read file: %s: %v", b.FileName, err)
	}
	return string(bytes)
}
