package rest

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/arvindpadev/MostPostsMostUpvotes/pkg/log"
)

type redditRestResponsePayload struct {
	RedditResponsePayload
	RateLimitUsed      int
	RateLimitRemaining int
	RateLimitReset     int
	Date               time.Time
}

const rateLimitUsedHeader = "x-ratelimit-used"
const rateLimitRemainingHeader = "x-ratelimit-remaining"
const rateLimitResetHeader = "x-ratelimit-reset"
const Bearer = "Bearer"
const UserAgent = "User-Agent"

func restGet(ctx context.Context, requestPath string, queryParameters *url.Values) (*redditRestResponsePayload, error) {
	log.Debug.Println(fmt.Sprintf("begin restGet(%s, %v)", requestPath, queryParameters.Encode()))
	defer log.Debug.Println(fmt.Sprintf("end restGet(%s, %v)", requestPath, queryParameters.Encode()))
	if !queryParameters.Has("raw_json") {
		queryParameters.Add("raw_json", "1")
	}

	u := url.URL{
		Scheme:   "https",
		Host:     "oauth.reddit.com",
		Path:     requestPath,
		RawQuery: queryParameters.Encode(),
	}

	requestUrl := u.String()
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", requestUrl, nil)
	if err != nil {
		log.Warn.Println(fmt.Sprintf("request creation error %s - %v", requestUrl, err))
		return nil, err
	}

	userAgent := fmt.Sprintf("%s", ctx.Value(UserAgent))
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", Bearer, ctx.Value(Bearer)))
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		log.Warn.Println(fmt.Sprintf("request error %s - %v", requestUrl, err))
		return nil, err
	}

	if resp.StatusCode == 429 {
		log.Error.Println(fmt.Sprintf("restGet(%s, %v) with request %v", requestPath, queryParameters.Encode(), req))
		panic("429 received")
	} else {
		log.Debug.Println(fmt.Sprintf("restGet(%s, %v) statusCode: (%d) => %v", requestPath, queryParameters.Encode(), resp.StatusCode, resp))
	}

	defer resp.Body.Close()
	redditResponse := redditRestResponsePayload{}
	redditResponse.StatusCode = resp.StatusCode
	redditResponse.RateLimitUsed = rateLimitHeaderValue(rateLimitUsedHeader, requestUrl, &resp.Header)
	redditResponse.RateLimitRemaining = rateLimitHeaderValue(rateLimitRemainingHeader, requestUrl, &resp.Header)
	redditResponse.RateLimitReset = rateLimitHeaderValue(rateLimitResetHeader, requestUrl, &resp.Header)
	date := resp.Header.Get("date")
	redditResponse.Date, err = time.Parse(time.RFC1123, date)
	log.Debug.Println(fmt.Sprintf("restGet: statusCode(%d), redditResponse.RateLimitXXX{%s %d %d %d} %s", redditResponse.StatusCode, date, redditResponse.RateLimitRemaining, redditResponse.RateLimitReset, redditResponse.RateLimitUsed, requestUrl))
	if err != nil {
		redditResponse.Date = time.Now()
		log.Warn.Println(fmt.Sprintf("unable to parse date %s - %s --> %v", requestUrl, date, err))
	}

	redditResponse.Content, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Warn.Println(fmt.Sprintf("response body reading error %s - %v", requestUrl, err))
	}

	return &redditResponse, err
}

func rateLimitHeaderValue(headerName string, requestUrl string, headers *http.Header) int {
	log.Debug.Println(fmt.Sprintf("begin rateLimitHeaderValue(%s, %s, %v)", headerName, requestUrl, *headers))
	defer log.Debug.Println(fmt.Sprintf("end rateLimitHeaderValue(%s, %s, %v)", headerName, requestUrl, *headers))
	rateLimitValue := headers.Get(headerName)
	r, rateLimitErr := strconv.ParseFloat(rateLimitValue, 21)
	if rateLimitErr != nil {
		log.Warn.Println(fmt.Sprintf("incorrect %s header value %s - %s", headerName, requestUrl, rateLimitValue))
		r = -1
	}

	rateLimit := int(math.Round(r))
	return rateLimit
}
