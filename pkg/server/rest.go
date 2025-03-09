package server

import (
	"encoding/json"
	"fmt"
	golog "log"
	"net/http"
	"strings"

	"github.com/arvindpadev/MostPostsMostUpvotes/pkg/log"
	"github.com/arvindpadev/MostPostsMostUpvotes/pkg/posts"
)

type HttpServerResponse struct {
	ID36  []string `json:"id36"`
	Count int      `json:"count"`
}

func runHttpServer(port string, bearerTokenChannel <-chan string) {
	ladder := posts.SingleRankLadder()
	authorsMostPostsApi := "GET /authors/most_posts"
	postsMostVotesApi := "GET /posts/most_votes"
	mux := http.NewServeMux()
	mux.HandleFunc(authorsMostPostsApi, func(w http.ResponseWriter, _ *http.Request) {
		topAuthors, postCount := ladder.GetAuthorsWithMostPosts()
		log.Info.Println(fmt.Sprintf("Authors with mosts posts (%d) %v", postCount, topAuthors))
		writeResponse(w, topAuthors, postCount)
	})
	mux.HandleFunc(postsMostVotesApi, func(w http.ResponseWriter, _ *http.Request) {
		topPosts, voteCount := ladder.GetPostsWithMostVotes()
		log.Info.Println(fmt.Sprintf("Posts with most votes (%d) %v", voteCount, topPosts))
		writeResponse(w, topPosts, voteCount)
	})

	p := fmt.Sprintf(":%s", port)
	message := []string{
		fmt.Sprintf("Starting server on port %s", port),
		"Available APIs:",
		authorsMostPostsApi,
		postsMostVotesApi,
	}

	log.Info.Println(strings.Join(message, "\n"))
	golog.Fatal(http.ListenAndServe(p, mux))
}

func writeResponse(w http.ResponseWriter, id36 []string, count int) {
	if len(id36) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response := &HttpServerResponse{Count: count, ID36: id36}
	content, err := json.Marshal(response)
	if err != nil {
		log.Error.Println(fmt.Sprintf("Unable to marshal %v --> %v", *response, err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(content)
}
