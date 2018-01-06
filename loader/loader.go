package loader

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/binatify/simple-wrk/util"
)

type Loader struct {
	goroutines int
	duration   int
	testUrl    string

	interrupted int32

	summray *RequesterStats
}

func NewLoader(goroutines, duration int, testUrl string) *Loader {
	summary := &RequesterStats{
		MinRequestTime: time.Minute,
	}

	return &Loader{goroutines, duration, testUrl, 0, summary}
}

type TotalStats struct {
	RequestRate float64 //per second
	BytesRate   float64 // per second

	AvgThreadTime  time.Duration
	AvgRequestTime time.Duration

	*RequesterStats
}

type RequesterStats struct {
	SuccessRequests int
	ErrRequests     int
	TotRespSize     int64

	TotDuration    time.Duration
	MinRequestTime time.Duration
	MaxRequestTime time.Duration
}

func (this *Loader) Run() {
	var wg sync.WaitGroup

	statsChan := make(chan *RequesterStats, this.goroutines)

	for i := 0; i < this.goroutines; i++ {
		wg.Add(1)

		go func() {
			httpClient := NewClient(this.testUrl)

			stats := &RequesterStats{MinRequestTime: time.Minute}
			start := time.Now()

			for time.Since(start).Seconds() <= float64(this.duration) && atomic.LoadInt32(&this.interrupted) == 0 {
				respSize, reqDur := httpClient.DoRequest()
				if respSize > 0 {
					stats.TotRespSize += int64(respSize)
					stats.TotDuration += reqDur
					stats.MaxRequestTime = util.MaxDuration(reqDur, stats.MaxRequestTime)
					stats.MinRequestTime = util.MinDuration(reqDur, stats.MinRequestTime)
					stats.SuccessRequests++
				} else {
					stats.ErrRequests++
				}
			}
			statsChan <- stats
		}()
	}

	go func() {
		for stats := range statsChan {
			this.summray.ErrRequests += stats.ErrRequests
			this.summray.SuccessRequests += stats.SuccessRequests
			this.summray.TotRespSize += stats.TotRespSize
			this.summray.TotDuration += stats.TotDuration
			this.summray.MaxRequestTime = util.MaxDuration(this.summray.MaxRequestTime, stats.MaxRequestTime)
			this.summray.MinRequestTime = util.MinDuration(this.summray.MinRequestTime, stats.MinRequestTime)

			wg.Done()
		}
	}()

	wg.Wait()

	close(statsChan)
}

func (this *Loader) TotalStats() (totalStats *TotalStats) {
	totalStats = &TotalStats{
		RequesterStats: this.summray,
	}

	if this.summray.SuccessRequests == 0 {
		return
	}

	totalStats.AvgThreadTime = this.summray.TotDuration / time.Duration(this.goroutines)
	totalStats.RequestRate = float64(this.summray.SuccessRequests) / totalStats.AvgThreadTime.Seconds()
	totalStats.BytesRate = float64(this.summray.TotRespSize) / totalStats.AvgThreadTime.Seconds()
	totalStats.AvgRequestTime = this.summray.TotDuration / time.Duration(this.summray.SuccessRequests)

	return
}

func (this *Loader) Stop() {
	atomic.StoreInt32(&this.interrupted, 1)
}
