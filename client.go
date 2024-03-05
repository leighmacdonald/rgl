package rgl

import (
	"context"
	"errors"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

const (
	maxBucket     = 50
	limitInterval = 10 * time.Second
)

var (
	ErrRequestWait = errors.New("failed to wait for request")
)

type LimiterClient struct {
	*http.Client
	*rate.Limiter
}

func (c *LimiterClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if errWait := c.Wait(ctx); errWait != nil {
		return nil, errors.Join(errWait, ErrRequestWait)
	}

	resp, errDo := c.Client.Do(req)
	if errDo != nil {
		return nil, errors.Join(errDo, ErrRequestCreate)
	}

	return resp, nil
}

func NewClient() *LimiterClient {
	return &LimiterClient{
		Client:  &http.Client{Timeout: time.Second * 15},
		Limiter: rate.NewLimiter(rate.Every(limitInterval), maxBucket),
	}
}
