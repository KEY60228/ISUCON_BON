package main

import (
	"fmt"
	"strings"
	"time"
)

type Option struct {
	TargetHost               string
	RequestTimeout           time.Duration
	InitializeRequestTimeout time.Duration
	ExitErrorOnFail          bool
}

func (o Option) String() string {
	args := []string{
		"benchmarker",
		fmt.Sprintf("--target-host=%s", o.TargetHost),
		fmt.Sprintf("--request-timeout=%s", o.RequestTimeout.String()),
		fmt.Sprintf("--initialize-request-timeout=%s", o.InitializeRequestTimeout.String()),
		fmt.Sprintf("--exit-error-on-fail=%v", o.ExitErrorOnFail),
	}
	return strings.Join(args, " ")
}
