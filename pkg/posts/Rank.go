package posts

import (
	"fmt"

	"github.com/arvindpadev/MostPostsMostUpvotes/pkg/log"
)

type ladderEntry struct {
	count int
	name  string
}

type rankLadder struct {
	authors  []ladderEntry
	posts    []ladderEntry
	byPost   map[string]int
	byAuthor map[string]int
}

func NewRankLadder() *rankLadder {
	return &rankLadder{
		authors:  make([]ladderEntry, 0, 0),
		posts:    make([]ladderEntry, 0, 0),
		byPost:   make(map[string]int),
		byAuthor: make(map[string]int),
	}
}

func (r *rankLadder) rank(posts []Post) {
	log.Debug.Println(fmt.Sprintf("begin rank(%v)", posts))
	defer log.Debug.Println(fmt.Sprintf("end rank(%v)", posts))
	for _, post := range posts {
		r.posts = placeAtCorrectRank(post.Name, post.Upvotes, r.posts, &r.byPost)
		if _, ok := r.byPost[post.Name]; !ok {
			authorPostCount := 1
			if _, aok := r.byAuthor[post.Author]; aok {
				authorIndex := r.byAuthor[post.Author]
				authorPostCount = r.authors[authorIndex].count + 1
			}

			r.authors = placeAtCorrectRank(post.Author, authorPostCount, r.authors, &r.byAuthor)
		}
	}
}

func (r *rankLadder) getAuthorsWithMostPosts() ([]string, int) {
	return getTopEntries(r.authors)
}

func (r *rankLadder) getPostsWithMostVotes() ([]string, int) {
	return getTopEntries(r.posts)
}

func getTopEntries(ladder []ladderEntry) ([]string, int) {
	results := make([]string, 0)
	log.Debug.Println(fmt.Sprintf("begin getTopEntries(%v)", ladder))
	defer log.Debug.Println(fmt.Sprintf("begin getTopEntries(%v) returned %v", ladder, results))
	top := ladder[0].count
	for _, entry := range ladder {
		if entry.count < top {
			break
		}

		results = append(results, entry.name)
	}

	return results, top
}

func placeAtCorrectRank(name string, count int, ladder []ladderEntry, l *map[string]int) []ladderEntry {
	newEntry := ladderEntry{name: name, count: count}
	lookup := *l
	if len(ladder) == 0 {
		lookup[newEntry.name] = 0
		return append(ladder, newEntry)
	} else if _, ok := lookup[name]; ok {
		i := lookup[name]
		ladder[i].count = count
		for ; i < len(ladder)-1; i = i + 1 {
			if ladder[i].count < ladder[i+1].count {
				swap(ladder, l, i, i+1)
			} else {
				break
			}
		}

		for ; i > 0; i = i - 1 {
			if ladder[i].count > ladder[i-1].count {
				swap(ladder, l, i, i-1)
			} else {
				break
			}
		}
		return ladder
	}

	index := binarySearchIndex(count, ladder, len(ladder)-1)
	left := append(ladder[0:index], newEntry)
	newLadder := append(left, ladder[index:]...)
	for _, entry := range ladder[index:] {
		lookup[entry.name] = lookup[entry.name] + 1
	}

	return newLadder
}

func binarySearchIndex(count int, ladder []ladderEntry, end int) int {
	debugMsg := fmt.Sprintf("binarySearchIndex %d %d %v", end, count, ladder)
	log.Debug.Println(debugMsg)
	if end > 1 {
		mid := end / 2
		if count == ladder[mid].count {
			i := mid
			for ; i < end+1 && count == ladder[i].count; i = i + 1 {
			}
			log.Debug.Println(fmt.Sprintf("Return %d for %s", i, debugMsg))
			return i
		}

		slice := ladder[:mid]
		newStart := 0
		if count < ladder[mid].count {
			slice = ladder[mid+1:]
			newStart = mid + 1
		}

		retval := newStart + binarySearchIndex(count, slice, len(slice)-1)
		log.Debug.Println(fmt.Sprintf("Return %d upon recursion for %s", retval, debugMsg))
		return retval
	}

	for i := 0; i <= end; i = i + 1 {
		if count > ladder[i].count {
			log.Debug.Println(fmt.Sprintf("Return %d with end=%d for %s", i, end, debugMsg))
			return i
		}
	}

	return end + 1
}

func swap(ladder []ladderEntry, l *map[string]int, index1, index2 int) {
	lookup := *l
	lookup[ladder[index1].name] = index2
	lookup[ladder[index2].name] = index1
	tempName := ladder[index1].name
	tempCount := ladder[index1].count
	ladder[index1].name = ladder[index2].name
	ladder[index1].count = ladder[index2].count
	ladder[index2].name = tempName
	ladder[index2].count = tempCount
}
