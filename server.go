package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"net/http"
	"time"
)

type Server struct {
	codex *Codex
	addr string
	watcher *fsnotify.Watcher
}

func NewServer(paths []string, addr string) *Server {
	codex, err := NewCodex(paths)
	if err != nil {
		log.Fatal(err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	for _, codoc := range codex.inputs {
		if err = watcher.Add(codoc.path); err != nil {
			log.Fatal(err)
		}
	}

	return &Server{
		codex: codex,
		addr: addr,
		watcher: watcher,
	}
}

func (srv *Server) Start() {
	log.Println("Starting with", len(srv.codex.inputs), "input document(s)")
	if err := srv.codex.Build(); err != nil {
		log.Fatal(err)
	}
	log.Println("Finished building from", len(srv.codex.inputs), "docs")

	go srv.Watch()
	go srv.Serve()
	select {}
}

func (srv *Server) Watch() {
	log.Println("Watching", len(srv.codex.inputs), "docs for changes ...")
	for {
		select {
		case event, ok := <-srv.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				srv.OnFileChange(event.Name)
			}
		case err, ok := <-srv.watcher.Errors:
			if !ok {
				return
			}
			log.Println("watch error:", err)
		}
	}
}

// FIXME debounce
func (srv *Server) OnFileChange(path string) {
	log.Println(path, "has changed: rebuilding ...")
	if err := srv.codex.Build(); err != nil {
		log.Fatal(err)
	}
	log.Println("Finished rebuilding")

	html := srv.codex.Output()
}

func (srv *Server) Serve() {
	http.Handle("/static/", http.FileServer(http.Dir(RootDir())))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		fmt.Fprintf(w, srv.codex.Output())
	})

	log.Println("Starting server at address", srv.addr)
    if err := http.ListenAndServe(srv.addr, nil); err != nil {
        log.Fatal(err)
    }
}
