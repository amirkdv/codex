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

type Server struct {
	codex   *Codex
	addr    string
	watcher *fsnotify.Watcher

	websockets []*websocket.Conn
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
		codex:   codex,
		addr:    addr,
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
	debouncer := time.NewTimer(debounceWait)
	for {
		select {
		case event, ok := <-srv.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				// don't actually trigger the event handler, just set the timer
				log.Println(event.Name, "changed: triggering rebuild ...")
				debouncer.Reset(debounceWait)
			}
		case err, ok := <-srv.watcher.Errors:
			if !ok {
				return
			}
			log.Println("watch error:", err)
		case <-debouncer.C:
			// caution: the current debouncer assumes all inputs are reparsed on
			// any file change, regardless of which file. If this is optimized,
			// the debouncer needs to be more sophisticated.
			srv.OnFileChange()
		}
	}
}

func (srv *Server) OnFileChange() {
	if err := srv.codex.Build(); err != nil {
		log.Fatal(err)
	}
	log.Println("Finished rebuilding")

	html := srv.codex.Output()

	// update clients
	for idx, ws := range srv.websockets {
		log.Println("Updating", len(srv.websockets), "websocket(s)")
		if err := ws.WriteMessage(websocket.TextMessage, []byte(html)); err != nil {
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
	http.Handle("/static/", http.FileServer(http.Dir(RootDir())))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, srv.codex.Output())
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return // TODO when does this happen?
		}
		log.Println("Accepted new websocket from", r.RemoteAddr)
		srv.websockets = append(srv.websockets, ws)
	})

	log.Println("Starting server at address", srv.addr)
	if err := http.ListenAndServe(srv.addr, nil); err != nil {
		log.Fatal(err)
	}
}
