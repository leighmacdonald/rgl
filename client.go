package rgl

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/time/rate"
)

const (
	maxBucket     = 50
	limitInterval = 10 * time.Second
)

type LimiterClient struct {
	*http.Client
	*rate.Limiter
}

func (c *LimiterClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if errWait := c.Wait(ctx); errWait != nil {
		return nil, errors.Wrap(errWait, "Failed to wait for request")
	}

	resp, errDo := c.Client.Do(req)
	if errDo != nil {
		return nil, errors.Wrap(errDo, "Failed to make request")
	}

	return resp, nil
}

func NewClient() *LimiterClient {
	return &LimiterClient{
		Client:  http.DefaultClient,
		Limiter: rate.NewLimiter(rate.Every(limitInterval), maxBucket),
	}
}
