package main

import (
	"os"

	"github.com/arvindpadev/MostPostsMostUpvotes/pkg/server"
)

func main() {
	server.RankPosts(os.Args[1:])
}
