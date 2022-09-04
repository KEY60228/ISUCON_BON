package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/isucon/isucandar"
)

var (
	ContestantLogger = log.New(os.Stdout, "", log.Ltime|log.Lmicroseconds)
	AdminLogger      = log.New(os.Stderr, "[ADMIN]", log.Ltime|log.Lmicroseconds)
)

const (
	DefaultTargetHost               = "localhost:8080"
	DefaultRequestTimeout           = 3 * time.Second
	DefaultinitializeRequestTimeout = 10 * time.Second
	DefaultExitErrorOnFail          = true
)

func main() {
	var option Option

	flag.StringVar(&option.TargetHost, "target-host", DefaultTargetHost, "Benchmark target host with port")
	flag.DurationVar(&option.RequestTimeout, "request-timeout", DefaultRequestTimeout, "Default request timeout")
	flag.DurationVar(&option.InitializeRequestTimeout, "initialize-request-timeout", DefaultinitializeRequestTimeout, "Initialize request timeout")
	flag.BoolVar(&option.ExitErrorOnFail, "exit-error-on-fail", DefaultExitErrorOnFail, "Exit with error if benchmark fails")
	flag.Parse()

	AdminLogger.Print(option)

	scenario := &Scenario{Option: option}
	benchmark, err := isucandar.NewBenchmark(
		isucandar.WithoutPanicRecover(),
		isucandar.WithLoadTimeout(1*time.Minute),
	)
	if err != nil {
		AdminLogger.Fatal(err)
	}

	benchmark.AddScenario(scenario)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := benchmark.Start(ctx)

	for _, err := range result.Errors.All() {
		ContestantLogger.Printf("%v", err)
		AdminLogger.Printf("%+v", err)
	}
}
