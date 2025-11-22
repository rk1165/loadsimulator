package load

import (
	"context"
	"log"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rk1165/loadsimulator/internal/types"
)

type Stats struct {
	Success uint64
	Fail    uint64
	Total   uint64
	MinTime time.Duration
	AvgTime time.Duration
	P50     time.Duration
	P90     time.Duration
	P95     time.Duration
	P99     time.Duration
	MaxTime time.Duration
}

type Log struct {
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

type Load interface {
	Success(response any) bool
	CalculateStats() *Stats
	Execute(ctx context.Context, id uint64) error
}

type BaseLoad struct {
	Cfg           types.Config
	OK            atomic.Uint64
	KO            atomic.Uint64
	Total         atomic.Uint64
	ResponseTimes []time.Duration
	Mu            sync.Mutex
}

func NewBaseLoad(cfg types.Config) BaseLoad {
	expectedRequests := cfg.RatePerSec * cfg.Duration
	return BaseLoad{
		Cfg:           cfg,
		ResponseTimes: make([]time.Duration, 0, expectedRequests),
		Mu:            sync.Mutex{},
	}
}

func (b *BaseLoad) Record(responseTime time.Duration, ok bool) {
	b.Total.Add(1)
	if ok {
		b.OK.Add(1)
	} else {
		b.KO.Add(1)
	}
	b.Mu.Lock()
	b.ResponseTimes = append(b.ResponseTimes, responseTime)
	b.Mu.Unlock()
}

func (b *BaseLoad) CalculateStats() *Stats {
	total := b.Total.Load()
	ok := b.OK.Load()
	ko := b.KO.Load()
	if len(b.ResponseTimes) == 0 {
		return &Stats{
			Total:   total,
			Success: ok,
			Fail:    ko,
		}
	}

	sort.Slice(b.ResponseTimes, func(i, j int) bool {
		return b.ResponseTimes[i] < b.ResponseTimes[j]
	})

	var sum time.Duration
	for _, duration := range b.ResponseTimes {
		sum += duration
	}
	n := len(b.ResponseTimes)
	avg := sum / time.Duration(n)

	stats := &Stats{
		Total:   total,
		Success: ok,
		Fail:    ko,
		MinTime: b.ResponseTimes[0],
		P50:     b.ResponseTimes[n/2],
		AvgTime: avg,
		P90:     b.ResponseTimes[int(float64(n)*0.9)],
		P95:     b.ResponseTimes[int(float64(n)*0.95)],
		P99:     b.ResponseTimes[int(float64(n)*0.99)],
		MaxTime: b.ResponseTimes[n-1],
	}
	return stats
}
