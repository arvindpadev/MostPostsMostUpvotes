package posts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"runtime"
	"time"

	"github.com/arvindpadev/MostPostsMostUpvotes/pkg/log"
	"github.com/arvindpadev/MostPostsMostUpvotes/pkg/rest"
)

type poller struct {
	postsChannel chan<- []Post
	bearerToken  string
}

func PollPosts(username, password, secret, script string, bearerTokenChannel chan<- string) {
	postsChannel := make(chan []Post, 1000)
	defer close(postsChannel)
	p := &poller{
		postsChannel: postsChannel,
	}

	go p.pollAuth(username, password, secret, script, bearerTokenChannel)
	go p.pollGetPosts("top")
	go p.pollGetPosts("new")
	ranker := NewRankLadder()
	for {
		posts := <-postsChannel
		log.Debug.Println(fmt.Sprintf("ranker goroutine received %d posts [%v]", len(posts), posts))
		ranker.rank(posts)
		topAuthors, postCount := ranker.getAuthorsWithMostPosts()
		log.Info.Println(fmt.Sprintf("Authors with mosts posts (%d) %v", postCount, topAuthors))
		topPosts, voteCount := ranker.getPostsWithMostVotes()
		log.Info.Println(fmt.Sprintf("Posts with most votes (%d) %v", voteCount, topPosts))
	}
}

func (p *poller) pollGetPosts(path string) {
	after := ""
	for {
		log.Debug.Println(fmt.Sprintf("pollGetPosts Bearer Token %s", p.bearerToken))
		if len(p.bearerToken) > 0 {
			log.Debug.Println(fmt.Sprintf("pollGetPosts { path: %s, after: %s } ready", path, after))
			ctx := context.WithValue(context.Background(), rest.Bearer, p.bearerToken)
			ctx = context.WithValue(ctx, rest.UserAgent, fmt.Sprintf("%s:JhApCodingChallenge:v0.1.0 (by u/ToyDev)", runtime.GOOS))
			mostRecentPosts := p.getPosts(ctx, path, after)
			p.postsChannel <- mostRecentPosts.posts
			if mostRecentPosts.backoff > 0 {
				log.Debug.Println(fmt.Sprintf("pollGetPosts %s sleeping for %d seconds", path, mostRecentPosts.backoff))
				time.Sleep(time.Duration(mostRecentPosts.backoff) * time.Second)
			} else {
				after = mostRecentPosts.after
			}
		}
	}
}

func (p *poller) pollAuth(username, password, secret, script string, bearerTokenChan chan<- string) {
	for {
		log.Debug.Println("Bearer token goroutine starting")
		client := rest.NewRedditAuthClient(username, password, secret, script)
		bt, expiresIn := client.Authenticate()
		p.bearerToken = bt
		bearerTokenChan <- p.bearerToken
		if expiresIn < 100 {
			expiresIn = 3600
		}

		log.Debug.Println(fmt.Sprintf("Bearer token goroutine sleeping for %d seconds", expiresIn-100))
		sleepDuration := time.Duration(expiresIn - 100)
		time.Sleep(sleepDuration * time.Second)
	}
}

func (p *poller) getPosts(ctx context.Context, path string, after string) *mostRecentPosts {
	log.Debug.Println(fmt.Sprintf("begin getPosts %s %s", path, after))
	client := rest.SingleRedditRestClient()
	values := url.Values{}
	if len(after) > 0 {
		values.Add("after", after)
	}

	mostRecentPosts := mostRecentPosts{
		posts: make([]Post, 0),
		after: after,
	}

	response := client.Get(ctx, path, &values)
	mostRecentPosts.backoff = response.Backoff
	if response.ResponseError == nil && response.Payload.StatusCode == 200 {
		var listingResponse redditListingResponse
		err := json.Unmarshal(response.Payload.Content, &listingResponse)
		if err == nil {
			mostRecentPosts.after = listingResponse.Listing.After
			for _, child := range listingResponse.Listing.Children {
				if child.Kind == "t3" {
					mostRecentPosts.posts = append(mostRecentPosts.posts, child.RedditPost)
				}
			}
		} else {
			log.Warn.Println(fmt.Sprintf("Unmarshal error getPosts %s, %v", after, err))
		}
	}

	return &mostRecentPosts
}
