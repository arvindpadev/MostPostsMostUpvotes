package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/arvindpadev/MostPostsMostUpvotes/pkg/log"
)

type RedditAuthClient interface {
	Authenticate() (string, int)
}

type redditAuthPostResponseBody struct {
	Token     string `json:"access_token"`
	Type      string `json:"token_type"`
	ExpiresIn int    `json:"expires_in"`
	Scope     string `json:"scope"`
}

type redditBearerAuthClient struct {
	postBody string
	secret   string
	script   string
}

func NewRedditAuthClient(username string, password string, secret string, script string) RedditAuthClient {
	log.Debug.Println("begin NewRedditAuthClient")
	defer log.Debug.Println("end NewRedditAuthClient")
	values := url.Values{}
	values.Add("grant_type", "client_credentials")
	values.Add("username", username)
	values.Add("password", password)
	postBody := values.Encode()
	log.Debug.Println("NewRedditAuthClient postBody= " + postBody)
	return &redditBearerAuthClient{secret: secret, script: script, postBody: string(postBody)}
}

func (r *redditBearerAuthClient) Authenticate() (string, int) {
	log.Debug.Println("begin redditBearerAuthClient.Authenticate()")
	defer log.Debug.Println("end redditBearerAuthClient.Authenticate()")
	reader := strings.NewReader(r.postBody)
	req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", reader)
	if err != nil {
		log.Error.Println(fmt.Sprintf("authentication request creation error %v", err))
		panic("authentication http request failure")
	}

	req.SetBasicAuth(r.script, r.secret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "JhApCodingChallenge")
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error.Println(fmt.Sprintf("authentication response error %v for request %v", err, *req))
		panic("authentication http response error")
	}

	if resp.StatusCode == 429 {
		return "", 0
	}

	if resp.StatusCode != 200 {
		log.Error.Println(fmt.Sprintf("Authentication error %v for request %v", *resp, *req))
		panic("authentication failure")
	}

	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error.Println(fmt.Sprintf("Authentication response body read error %v", err))
		panic("authentication response reading failure")
	}

	var response redditAuthPostResponseBody
	err = json.Unmarshal(content, &response)
	if err != nil {
		log.Error.Println(fmt.Sprintf("Failed to unmarshal auth response body %v", err))
		panic("Auth unmarshal failure")
	}

	return response.Token, response.ExpiresIn
}
