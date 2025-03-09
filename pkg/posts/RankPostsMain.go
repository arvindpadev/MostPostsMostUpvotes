package posts

import (
	"fmt"

	"strings"

	"github.com/arvindpadev/MostPostsMostUpvotes/pkg/log"
)

func RankPosts(args []string) {
	logLevel := "info"
	var script, secret, username, password string
	for i := 0; i < len(args)-1; i = i + 2 {
		switch strings.ToLower(args[i]) {
		case "--help":
			printUsage()
			return
		case "help":
			printUsage()
			return
		case "--username":
			username = args[i+1]
		case "--password":
			password = args[i+1]
		case "--script":
			script = args[i+1]
		case "--secret":
			secret = args[i+1]
		case "--loglevel":
			logLevel = args[i+1]
		default:
			panic(fmt.Sprintf("Bad argument %s received in command line arguments %v", args[i], args))
		}
	}

	log.InitLoggers(logLevel)
	var bearerToken string
	bearerTokenChan := make(chan string, 1)
	defer close(bearerTokenChan)
	go func(bearerToken *string, bearerTokenChannel <-chan string) {
		for {
			*bearerToken = <-bearerTokenChannel
		}
	}(&bearerToken, bearerTokenChan)
	PollPosts(username, password, secret, script, bearerTokenChan)
}

func printUsage() {
	fmt.Println("USAGE: ./cmd --script <reddit script> --secret <reddit secret> --username <reddit username> --password <reddit password>")
	fmt.Println("HELP: './cmd --help' OR './cmd help' shows this text")
	fmt.Println("To set up a script and secret, please take a look at https://github.com/reddit-archive/reddit/wiki/OAuth2")
}
