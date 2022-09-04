package main

import (
	"context"
	"net/http"

	"github.com/isucon/isucandar/agent"
)

func GetInitializeAction(ctx context.Context, ag *agent.Agent) (*http.Response, error) {
	req, err := ag.GET("/initialize")
	if err != nil {
		return nil, err
	}
	return ag.Do(ctx, req)
}
