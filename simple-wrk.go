package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	wrk := loader.NewLoader(goroutines, duration, testUrl)

	go func() {
		select {
		case <-sigChan:
			wrk.Stop()
			fmt.Printf("stopping...\n")
		}
	}()

	wrk.Run()

	totalStats := wrk.TotalStats()

	if totalStats.SuccessRequests == 0 {
		fmt.Println("Error: No statistics collected / no requests found")
		return
	}

	fmt.Printf("%v requests in %v, %v read\n",
		totalStats.SuccessRequests,
		totalStats.AvgThreadTime,
		util.ByteSize(float64(totalStats.TotRespSize)))

	fmt.Printf("Requests/sec:\t\t%.2f\nTransfer/sec:\t\t%v\nAvg Req Time:\t\t%v\n",
		totalStats.RequestRate,
		util.ByteSize(totalStats.BytesRate),
		totalStats.AvgRequestTime)

	fmt.Printf("Fastest Request:\t%v\n", totalStats.MinRequestTime)
	fmt.Printf("Slowest Request:\t%v\n", totalStats.MaxRequestTime)
	fmt.Printf("Number of Errors:\t%v\n", totalStats.ErrRequests)
}
