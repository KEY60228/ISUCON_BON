package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/isucon/isucandar/agent"
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

func (o Option) NewAgent(forInitialize bool) (*agent.Agent, error) {
	agentOptions := []agent.AgentOption{
		agent.WithBaseURL(fmt.Sprintf("http://%s/", o.TargetHost)),
		agent.WithCloneTransport(agent.DefaultTransport),
	}

	if forInitialize {
		agentOptions = append(agentOptions, agent.WithTimeout(o.InitializeRequestTimeout))
	} else {
		agentOptions = append(agentOptions, agent.WithTimeout(o.RequestTimeout))
	}

	return agent.NewAgent(agentOptions...)
}
