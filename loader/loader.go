package loader

import (
	"sync/atomic"
	"time"

	"github.com/binatify/simple-wrk/util"
)

type Loader struct {
	goroutines int
	duration   int
	testUrl    string

	statsAggregator chan *RequesterStats
	interrupted     int32
}

func NewLoader(goroutines, duration int, testUrl string, statsAggregator chan *RequesterStats) *Loader {
	return &Loader{goroutines, duration, testUrl, statsAggregator, 0}
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
	for i := 0; i < this.goroutines; i++ {
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
			this.statsAggregator <- stats
		}()
	}
}

func (this *Loader) Stop() {
	atomic.StoreInt32(&this.interrupted, 1)
}
