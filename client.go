package rgl

import (
	"context"
	"golang.org/x/time/rate"
	"net/http"
	"time"
)

var client *limiterClient

func init() {
	client = &limiterClient{
		Client:  http.DefaultClient,
		Limiter: rate.NewLimiter(rate.Every(10*time.Second), 50),
	}
}

type limiterClient struct {
	*http.Client
	*rate.Limiter
}

func (c *limiterClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if errWait := c.Wait(ctx); errWait != nil {
		return nil, errWait
	}
	resp, errDo := c.Client.Do(req)
	if errDo != nil {
		return nil, errDo
	}
	return resp, nil
}
