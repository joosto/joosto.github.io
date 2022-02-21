package post

import (
	"bufio"
	"encoding/json"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/russross/blackfriday/v2"

	"joosto.github.io/internal/must"
)

type Post struct {
	Slug    string        `json:"slug"`
	Date    string        `json:"date"`
	Title   string        `json:"title"`
	Summary string        `json:"summary"`
	URL     string        `json:"url"`
	HTML    template.HTML `json:"-"`
}

func Parse() []Post {
	filePaths, err := filepath.Glob("md/posts/*.md")
	must.Succeed(err)
	posts := make([]Post, len(filePaths))
	for i, path := range filePaths {
		posts[i] = parse(path)
	}
	sort.Slice(posts, func(i, j int) bool {
		return filePaths[i] > filePaths[j]
	})
	return posts
}

func parse(path string) Post {
	add := func(b []byte, s string) []byte {
		return append(b, []byte(s+"\n")...)
	}
	file, err := os.Open(path)
	must.Succeed(err)
	scanner := bufio.NewScanner(file)
	var header, body []byte
	scanningHeader := true
	for scanner.Scan() {
		line := scanner.Text()
		if line == "```json" {
			continue // skip first line
		}
		if line == "```" {
			scanningHeader = false
			continue
		}
		if scanningHeader {
			header = add(header, scanner.Text())
		} else {
			body = add(body, scanner.Text())
		}
	}
	if header == nil {
		log.Fatal("invalid post: missing header")
	}
	var post Post
	must.Succeed(json.Unmarshal(header, &post))
	if body != nil {
		output := blackfriday.Run(body)
		post.HTML = template.HTML(output)
	}
	return post
}
