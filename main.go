package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"joosto.github.io/internal/must"

	"github.com/fsnotify/fsnotify"

	"joosto.github.io/internal/post"
)

type PageData struct {
	Type string
}

type IndexData struct {
	PageData
	Posts []post.Post
}

type PostData struct {
	PageData
	Post post.Post
}

func main() {
	buildTemplates()
	go watchForChanges()
	runServer()
}

func watchForChanges() {
	watcher, err := fsnotify.NewWatcher()
	must.Succeed(err)
	defer func() {
		must.Succeed(watcher.Close())
	}()
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("fsnotify event: ", event)
				if event.Op != fsnotify.Write {
					log.Println("dropping event")
					continue
				}
				buildTemplates()
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("fsnotify error: ", err)
			}
		}
	}()
	must.Succeed(watcher.Add("templates/"))
	must.Succeed(watcher.Add("md/posts"))
	<-done
}

func buildTemplates() {
	posts := post.Parse()
	templates := template.Must(template.ParseGlob("templates/*"))
	log.Println("building templates")
	file, err := os.Create("index.html")
	must.Succeed(err)
	err = templates.ExecuteTemplate(file, "index.tmpl", IndexData{
		PageData: PageData{Type: "index"},
		Posts:    posts,
	})
	for _, p := range posts {
		file, err := os.Create(fmt.Sprintf("posts/%s.html", p.Slug))
		must.Succeed(err)
		err = templates.ExecuteTemplate(file, "post.tmpl", PostData{
			PageData: PageData{Type: "post"},
			Post:     p,
		})
	}
	must.Succeed(err)
}

func runServer() {
	must.Succeed(http.ListenAndServe(":1380", http.FileServer(http.Dir("."))))
}
