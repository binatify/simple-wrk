package loader

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/binatify/simple-wrk/util"
)

type Loader struct {
	goroutines      int
	duration        int
	testUrl         string
	statsAggregator chan *RequesterStats

	interrupted int32
}

func NewLoader(goroutines, duration int, testUrl string, statsAggregator chan *RequesterStats) *Loader {
	return &Loader{goroutines, duration, testUrl, statsAggregator, 0}
}

type RequesterStats struct {
	TotRespSize    int64
	TotDuration    time.Duration
	MinRequestTime time.Duration
	MaxRequestTime time.Duration
	NumRequests    int
	NumErrs        int
}

func (this *Loader) Run() {
	stats := &RequesterStats{MinRequestTime: time.Minute}
	start := time.Now()

	httpClient := newClient()

	for time.Since(start).Seconds() <= float64(this.duration) && atomic.LoadInt32(&this.interrupted) == 0 {
		respSize, reqDur := doRequest(httpClient, this.testUrl)
		if respSize > 0 {
			stats.TotRespSize += int64(respSize)
			stats.TotDuration += reqDur
			stats.MaxRequestTime = util.MaxDuration(reqDur, stats.MaxRequestTime)
			stats.MinRequestTime = util.MinDuration(reqDur, stats.MinRequestTime)
			stats.NumRequests++
		} else {
			stats.NumErrs++
		}
	}
	this.statsAggregator <- stats
}

func (this *Loader) Stop() {
	atomic.StoreInt32(&this.interrupted, 1)
}

func doRequest(httpClient *http.Client, loadUrl string) (respSize int, duration time.Duration) {
	respSize, duration = -1, -1

	req, err := http.NewRequest(http.MethodGet, loadUrl, nil)
	if err != nil {
		fmt.Println("An error occured doing request", err)
		return
	}

	start := time.Now()
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("An error occured doing request", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("An error occured reading body", err)
		return
	}

	duration = time.Since(start)
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		respSize = len(body) + int(util.EstimateHttpHeadersSize(resp.Header))
	} else if resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusTemporaryRedirect {
		respSize = int(resp.ContentLength) + int(util.EstimateHttpHeadersSize(resp.Header))
	} else {
		fmt.Println("received status code", resp.StatusCode, "from", resp.Header, "content", string(body), req)
	}

	return
}
