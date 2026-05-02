package metrics

import "sync"

type Metrics struct {
	mu        sync.Mutex
	Processed int
	Failed    int
	Recovered int
}

var M = &Metrics{}

func IncrementProcessed() {
	M.mu.Lock()
	defer M.mu.Unlock()
	M.Processed++
}

func IncrementFailed() {
	M.mu.Lock()
	defer M.mu.Unlock()
	M.Failed++
}

func IncrementRecovered(count int) {
	M.mu.Lock()
	defer M.mu.Unlock()
	M.Recovered += count
}

func Snapshot() Metrics {
	M.mu.Lock()
	defer M.mu.Unlock()

	return Metrics{
		Processed: M.Processed,
		Failed:    M.Failed,
		Recovered: M.Recovered,
	}
}
