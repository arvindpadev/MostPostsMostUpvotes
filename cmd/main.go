package main

import (
	"os"

	"github.com/arvindpadev/MostPostsMostUpvotes/pkg/posts"
)

func main() {
	posts.RankPosts(os.Args[1:])
}
