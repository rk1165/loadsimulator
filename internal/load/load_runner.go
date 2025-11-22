package load

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rk1165/loadsimulator/internal/types"
)

const MaxConcurrency = 100

// Runner runs a task at a fixed rate (requests per second) for a duration
// Concurrency >= RPS * averageLatencySeconds (Little's Law)
type Runner struct {
	Load           Load
	Cfg            types.Config
	scheduledCount uint64
	startedCount   uint64
	completedCount uint64
	failedCount    uint64
	firstErr       atomic.Value
}

func NewLoadRunner(load Load, cfg types.Config) *Runner {
	return &Runner{Load: load, Cfg: cfg}
}

// Run executes the Execute method of a load at the configured rate for the given duration
// Returns earliest error (if any)
func (r *Runner) Run(ctx context.Context, statCh chan<- *Stats) error {
	if err := r.ValidateConfig(); err != nil {
		return err
	}
	cfg := r.Cfg
	cfg.InfoLog.Printf("[INIT LOAD CONFIG] rps=%d duration=%d concurrency=%d jitter=%s", cfg.RatePerSec, cfg.Duration, cfg.Concurrency, cfg.Jitter)

	// loadCh receives request by the scheduler to execute a load every 'firing' second
	// where 'firing' second is calculated based on rps and duration and jitter (if any)
	loadCh := make(chan time.Time, cfg.Concurrency)

	var wg sync.WaitGroup
	simulationStartTime := time.Now()
	r.StartWorkers(ctx, loadCh, &wg)

	go r.StartScheduler(ctx, loadCh, simulationStartTime)

	wg.Wait()
	cfg.InfoLog.Printf("[SUMMARY] scheduled=%d started=%d completed=%d failures=%d duration=%s",
		atomic.LoadUint64(&r.scheduledCount),
		atomic.LoadUint64(&r.startedCount),
		atomic.LoadUint64(&r.completedCount),
		atomic.LoadUint64(&r.failedCount),
		time.Since(simulationStartTime).Truncate(time.Millisecond))
	cfg.InfoLog.Println("-------------------------------------------------------------------------")
	if v := r.firstErr.Load(); v != nil {
		return fmt.Errorf("%w", v.(error))
	}
	cfg.InfoLog.Println("Load Run completed successfully")
	statCh <- r.Load.CalculateStats()
	close(statCh)
	return nil
}

// StartWorkers starts cfg.Concurrency number of workers for executing the load received on loadCh
func (r *Runner) StartWorkers(ctx context.Context, loadCh <-chan time.Time, wg *sync.WaitGroup) {
	cfg := r.Cfg
	var errOnce sync.Once

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func(workerID string) {
			defer wg.Done()
			for scheduled := range loadCh {
				started := time.Now()
				offset := started.Sub(scheduled) // difference between scheduled and started time
				requestId := atomic.AddUint64(&r.startedCount, 1)
				cfg.InfoLog.Printf("workerID=[%s] [START] requestID=%d offset=%s goroutines=%d",
					workerID, requestId, offset, runtime.NumGoroutine())
				if e := r.Load.Execute(ctx, requestId); e != nil {
					errOnce.Do(func() { r.firstErr.Store(e) })
					atomic.AddUint64(&r.failedCount, 1)
					cfg.ErrorLog.Printf("workerId=[%s] [FAIL] requestId=%d err=[%v]", workerID, requestId, e)
				} else {
					cfg.InfoLog.Printf("workerId=[%s] [DONE] requestId=%d elapsed=%s", workerID, requestId, time.Since(started))
				}
				atomic.AddUint64(&r.completedCount, 1)
			}
		}(fmt.Sprintf("%s-%d", cfg.Name, i))
	}
}

// StartScheduler starts a scheduler for scheduling the load on the workers
func (r *Runner) StartScheduler(ctx context.Context, loadCh chan time.Time, simulationStartTime time.Time) {
	cfg := r.Cfg
	interval := time.Second / time.Duration(cfg.RatePerSec) // frequency of firing requests
	totalRequests := cfg.RatePerSec * cfg.Duration
	defer close(loadCh)
	n := 0
	next := simulationStartTime
	//rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for n < totalRequests {
		fire := next
		//if cfg.Jitter > 0 {
		//	j := time.Duration(rnd.Int63n(int64(cfg.Jitter)))
		//	if rnd.Intn(2) == 0 {
		//		fire = fire.Add(-j)
		//	} else {
		//		fire = fire.Add(j)
		//	}
		//}
		if d := time.Until(fire); d > 0 {
			time.Sleep(d)
		}
		select {
		case loadCh <- fire:
			n++
			atomic.AddUint64(&r.scheduledCount, 1)
			next = simulationStartTime.Add(time.Duration(n) * interval)
			if n%cfg.RatePerSec == 0 {
				elapsed := time.Since(simulationStartTime)
				cfg.InfoLog.Printf("[PROGRESS] elapsed=%s scheduled=%d started=%d completed=%d failures=%d",
					elapsed.Truncate(time.Millisecond),
					atomic.LoadUint64(&r.scheduledCount),
					atomic.LoadUint64(&r.startedCount),
					atomic.LoadUint64(&r.completedCount),
					atomic.LoadUint64(&r.failedCount))
			}
		case <-ctx.Done():
			cfg.InfoLog.Printf("[SCHEDULER] stop: context cancelled after %d requests", n)
			return
		}
	}
	cfg.InfoLog.Printf("[SCHEDULER] completed: scheduled=%d requests in %s",
		n, time.Since(simulationStartTime).Truncate(time.Millisecond))
}

func (r *Runner) ValidateConfig() error {
	cfg := r.Cfg
	if cfg.RatePerSec <= 0 {
		return errors.New("RPS must be > 0")
	}
	if cfg.Duration <= 0 {
		return errors.New("duration must be > 0")
	}
	if cfg.Concurrency <= 0 {
		return errors.New("concurrency must be > 0")
	}
	if cfg.Concurrency > MaxConcurrency {
		return errors.New("concurrency unreasonably high")
	}
	return nil
}
