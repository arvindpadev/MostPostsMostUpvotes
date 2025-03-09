package posts

type Post struct {
	Author  string `json:"author_fullname"`
	Name    string `json:"name"`
	Upvotes int    `json:"ups"`
}

type redditListingChild struct {
	Kind       string `json:"kind"`
	RedditPost Post   `json:"data"`
}

type redditListing struct {
	After    string               `json:"after"`
	Children []redditListingChild `json:"children"`
	Before   string               `json:"before"`
}

type redditListingResponse struct {
	Kind    string        `json:"kind"`
	Listing redditListing `json:"data"`
}

type mostRecentPosts struct {
	posts   []Post
	after   string
	backoff int64
}
