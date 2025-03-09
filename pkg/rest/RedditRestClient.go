package rest

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/arvindpadev/MostPostsMostUpvotes/pkg/log"
)

type RedditResponsePayload struct {
	StatusCode int
	Content    []byte
}

type RedditRestResponse struct {
	Payload       RedditResponsePayload
	ResponseError error
	Backoff       int64
}

type RedditRestClient interface {
	Get(ctx context.Context, requestUri string, queryParameters *url.Values) *RedditRestResponse
}

type redditRateLimitedClient struct {
	rateLimitReset             int
	rateLimitRemaining         int
	rateLimitResetReceivedTime time.Time
	mutex                      sync.RWMutex
}

var rateLimitedClient *redditRateLimitedClient = nil

func SingleRedditRestClient() RedditRestClient {
	log.Debug.Println("begin SingleRedditRestClient")
	defer log.Debug.Println("end SingleRedditRestClient")

	if rateLimitedClient == nil {
		log.Debug.Println("SingleRedditRestClient not previously created.")
		rateLimitedClient = &redditRateLimitedClient{}
		rateLimitedClient.rateLimitRemaining = 100
		rateLimitedClient.rateLimitReset = 60
		rateLimitedClient.rateLimitResetReceivedTime = time.Now()
		rateLimitedClient.mutex = sync.RWMutex{}
	}

	return rateLimitedClient
}

func (r *redditRateLimitedClient) Get(ctx context.Context, requestUri string, queryParameters *url.Values) *RedditRestResponse {
	log.Debug.Println(fmt.Sprintf("begin RateLimitedClient.Get(%s)", requestUri))
	defer log.Debug.Println(fmt.Sprintf("end RateLimitedClient.Get(%s)", requestUri))

	redditResponse := RedditRestResponse{
		Payload: RedditResponsePayload{StatusCode: 0, Content: []byte{}},
		Backoff: 0,
	}
	if r.rateLimitRemaining == 0 || r.rateLimitRemaining == -1 {
		backoff := r.getBackoff()
		if backoff > 0 {
			redditResponse.Backoff = int64(backoff)
			return &redditResponse
		}
	}

	payload, err := restGet(ctx, requestUri, queryParameters)
	if payload.Date.UnixNano() > r.rateLimitResetReceivedTime.UnixNano() {
		r.recordRateLimit(payload)
	}

	redditResponse.ResponseError = err
	redditResponse.Payload.StatusCode = payload.StatusCode
	redditResponse.Payload.Content = payload.Content
	return &redditResponse
}

func (r *redditRateLimitedClient) recordRateLimit(payload *redditRestResponsePayload) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	log.Debug.Println(fmt.Sprintf("begin RateLimitedClient{%v, %d, %d}.recordRateLimit(%v)", r.rateLimitResetReceivedTime, r.rateLimitRemaining, r.rateLimitReset, *payload))
	defer log.Debug.Println(fmt.Sprintf("end RateLimitedClient{%v}.recordRateLimit(%v)", r.rateLimitResetReceivedTime, *payload))

	if payload.Date.UnixNano() > r.rateLimitResetReceivedTime.UnixNano() {
		r.rateLimitResetReceivedTime = payload.Date
		r.rateLimitRemaining = payload.RateLimitRemaining
		r.rateLimitReset = payload.RateLimitReset
	}
}

func (r *redditRateLimitedClient) getBackoff() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	log.Debug.Println(fmt.Sprintf("begin RateLimitedClient{%v, %d, %d}.getBackoff()", r.rateLimitResetReceivedTime, r.rateLimitRemaining, r.rateLimitReset))
	defer log.Debug.Println(fmt.Sprintf("end RateLimitedClient{%v, %d, %d}.getBackoff()", r.rateLimitResetReceivedTime, r.rateLimitRemaining, r.rateLimitReset))
	if r.rateLimitRemaining == 0 || r.rateLimitRemaining == -1 {
		rateLimitReset := 60
		if r.rateLimitReset != 0 && r.rateLimitReset != -1 {
			rateLimitReset = r.rateLimitReset
		}

		durationSinceReceivedTime := time.Since(r.rateLimitResetReceivedTime)
		timeRemaining := time.Duration(rateLimitReset * 1000000000)
		if durationSinceReceivedTime < timeRemaining {
			return rateLimitReset
		}
	}

	return 0
}
