package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const (
	debounceWait = 200 * time.Millisecond
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

// Server is responsible for all long-running aspects of Codex:
//  - watching files for changes
//  - rebuilding DOM as pieces of it change,
//  - serving contents and static files over HTTP
//  - managing WebSocket connections for incremental updates.
//
// Concurrency model:
//  1. The first build on codex boot consumes all inputs in parallel, upto a
//     maximum concurrency level, see Codex.BuildAll()
//  2. Each subsequent build is triggered by a single file change, incremental
//     builds are always serialized; no two updates happen concurrently.
type Server struct {
	Codex *Codex
	Addr  string // whatever http.Listen() accepts

	watcher *fsnotify.Watcher
	status  map[string]string

	updates chan *Document
	builds  chan *Document

	websockets []*websocket.Conn
}

func NewServer(paths []string, addr string) *Server {
	cdx, err := NewCodex(paths)
	if err != nil {
		log.Fatal(err)
	}

	watcher, err := createWatcher(cdx.Inputs)
	if err != nil {
		log.Fatal(err)
	}

	return &Server{
		Codex:   cdx,
		Addr:    addr,
		watcher: watcher,
		status:  make(map[string]string),
		updates: make(chan *Document),
		builds:  make(chan *Document),
	}
}

func createWatcher(codocs map[string]*Document) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	for _, codoc := range codocs {
		if err := watcher.Add(codoc.Path); err != nil {
			return nil, err
		}
	}
	return watcher, nil
}

func (srv *Server) Start() {
	go srv.Watch()
	go srv.UpdateOnChange()
	go srv.Serve()
	select {}
}

func (srv *Server) Watch() {
	log.Println("Watching", len(srv.Codex.Inputs), "docs for changes ...")
	for {
		select {
		case event, ok := <-srv.watcher.Events:
			if !ok {
				log.Fatal("filesystem watcher crash!")
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				srv.updates <- srv.Codex.Inputs[event.Name]
			}
		case err, ok := <-srv.watcher.Errors:
			if !ok {
				log.Fatal("filesystem watcher crash!")
			}
			log.Println("watch error:", err)
		}
	}
}

func (srv *Server) UpdateOnChange() {
	for {
		select {
		case codoc := <-srv.updates:
			time.AfterFunc(debounceWait, func() {
				srv.builds <- codoc
			})
		case codoc := <-srv.builds:
			if codoc.CheckMtime().After(codoc.Btime) {
				log.Println("building:", codoc.Path)
				htmlStr, err := srv.Codex.Update(codoc)
				if err != nil {
					log.Fatal(err)
				}
				srv.UpdateClients(htmlStr)
			}
		}
	}
}

func (srv *Server) UpdateClients(htmlStr string) {
	for idx, ws := range srv.websockets {
		log.Println("Updating", len(srv.websockets), "websocket(s)")
		if err := ws.WriteMessage(websocket.TextMessage, []byte(htmlStr)); err != nil {
			log.Println("Failed to write to websocket,", err)
			srv.dropWebSocket(idx)
		}
	}
}

func (srv *Server) dropWebSocket(idx int) {
	log.Println("Dropping stale websocket:", srv.websockets[idx].RemoteAddr())
	nsocks := len(srv.websockets)
	srv.websockets[idx] = srv.websockets[nsocks-1]
	srv.websockets = srv.websockets[:nsocks-1]
}

func (srv *Server) Serve() {
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		static := STATICS[r.URL.Path]
		// note: this doesn't populate Content-Length which is mandatory!
		w.Header().Set("Content-Type", static.ContentType)
		fmt.Fprintf(w, static.Body)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, srv.Codex.Html())
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return // TODO when does this happen?
		}
		log.Println("Accepted new websocket from", r.RemoteAddr)
		srv.websockets = append(srv.websockets, ws)
	})

	log.Println("Starting server at address", srv.Addr)
	if err := http.ListenAndServe(srv.Addr, nil); err != nil {
		log.Fatal(err)
	}
}
