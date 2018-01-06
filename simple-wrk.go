package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/binatify/simple-wrk/loader"
	"github.com/binatify/simple-wrk/util"
)

var (
	goroutines, duration int
	testUrl              string
)

func init() {
	flag.IntVar(&goroutines, "c", 10, "Number of goroutines to use (concurrent connections)")
	flag.IntVar(&duration, "d", 5, "Duration of test in seconds")
}

func printDefaults() {
	fmt.Println("Usage: simple-wrk <options> <url>")
	fmt.Println("Options:")
	flag.VisitAll(func(flag *flag.Flag) {
		fmt.Println("\t-"+flag.Name, "\t", flag.Usage, "(Default "+flag.DefValue+")")
	})
}

func main() {
	flag.Parse()

	testUrl = flag.Arg(0)
	if testUrl == "" {
		printDefaults()
		os.Exit(1)
	}

	fmt.Printf("Running %ds test @ %s\n", duration, testUrl)

	statsAggregator := make(chan *loader.RequesterStats, goroutines)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	loadGen := loader.NewLoader(goroutines, duration, testUrl, statsAggregator)

	for i := 0; i < goroutines; i++ {
		go loadGen.Run()
	}

	responders := 0
	aggStats := loader.RequesterStats{MinRequestTime: time.Minute}

	for responders < goroutines {
		select {
		case <-sigChan:
			loadGen.Stop()
			fmt.Printf("stopping...\n")
		case stats := <-statsAggregator:
			aggStats.NumErrs += stats.NumErrs
			aggStats.NumRequests += stats.NumRequests
			aggStats.TotRespSize += stats.TotRespSize
			aggStats.TotDuration += stats.TotDuration
			aggStats.MaxRequestTime = util.MaxDuration(aggStats.MaxRequestTime, stats.MaxRequestTime)
			aggStats.MinRequestTime = util.MinDuration(aggStats.MinRequestTime, stats.MinRequestTime)
			responders++
		}
	}

	if aggStats.NumRequests == 0 {
		fmt.Println("Error: No statistics collected / no requests found")
		return
	}

	avgThreadDur := aggStats.TotDuration / time.Duration(responders)

	reqRate := float64(aggStats.NumRequests) / avgThreadDur.Seconds()
	avgReqTime := aggStats.TotDuration / time.Duration(aggStats.NumRequests)
	bytesRate := float64(aggStats.TotRespSize) / avgThreadDur.Seconds()

	fmt.Printf("%v requests in %v, %v read\n",
		aggStats.NumRequests,
		avgThreadDur,
		util.ByteSize(float64(aggStats.TotRespSize)))
	fmt.Printf("Requests/sec:\t\t%.2f\nTransfer/sec:\t\t%v\nAvg Req Time:\t\t%v\n",
		reqRate,
		util.ByteSize(bytesRate),
		avgReqTime)

	fmt.Printf("Fastest Request:\t%v\n", aggStats.MinRequestTime)
	fmt.Printf("Slowest Request:\t%v\n", aggStats.MaxRequestTime)
	fmt.Printf("Number of Errors:\t%v\n", aggStats.NumErrs)
}
