package main

import (
	"bytes"
	"os"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"path"
)

func RootDir() string {
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	return path.Dir(exe)
}

type Codex struct {
	inputs []*Codocument
	output *goquery.Document
}

func NewCodex(paths []string) (*Codex, error) {
	if len(paths) == 0 {
		return nil, errors.New("Need at least one input")
	}
	codocs := make([]*Codocument, len(paths))
	for idx, path := range paths {
		codocs[idx] = &Codocument{path: path}
	}
	return &Codex{inputs: codocs}, nil
}

func (cdx Codex) TransformAll() ([]*goquery.Document, error) {
	htmlDocs := make([]*goquery.Document, len(cdx.inputs))

	var errg errgroup.Group
	for idx, codoc := range cdx.inputs {
		// because closure around goroutine below
		idx := idx
		codoc := codoc
		errg.Go(func() error {
			doc, err := codoc.Transform()
			if err != nil {
				return err
			}
			htmlDocs[idx] = doc
			return nil
		})
	}
	if err := errg.Wait(); err != nil {
		return nil, err
	}
	return htmlDocs, nil
}

func (cdx *Codex) Build() error {
	// FIXME race
	docs, err := cdx.TransformAll()
	if err != nil {
		return err
	}

	outDoc, err := LoadHtml(CodexOutputTemplate)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer

	for _, doc := range docs {
		html, err := doc.Find("body").First().Html()
		if err != nil {
			return err
		}
		buffer.WriteString(html)
	}
	outDoc.Find("main").First().SetHtml(buffer.String())
	outDoc.Find(".node:not(:has(.node))").AddClass("node-leaf")

	cdx.output = outDoc
	return nil
}

func (cdx *Codex) BuildAndWatch() {
	if err := cdx.Build(); err != nil {
		log.Fatal(err)
	}
	log.Println("Finished building from", len(cdx.inputs), "docs")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go cdx.buildOnWrite(watcher)

	for _, codoc := range cdx.inputs {
		if err = watcher.Add(codoc.path); err != nil {
			log.Fatal(err)
		}
	}
	select {} // we're indefinitely waiting for fsnotify in separate goroutine
}

func (cdx *Codex) buildOnWrite(watcher *fsnotify.Watcher) {
	log.Println("Watching", len(cdx.inputs), "docs for changes ...")
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println(event.Name, "has changed: rebuilding ...")
				// FIXME debounce
				if err := cdx.Build(); err != nil {
					log.Fatal(err)
				}
				log.Println("Finished rebuilding")
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func (cdx *Codex) Serve(addr string) {
	http.Handle("/static/", http.FileServer(http.Dir(RootDir())))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		fmt.Fprintf(w, DocToHtml(cdx.output))
	})

	log.Println("Starting server at address", addr)
    if err := http.ListenAndServe(addr, nil); err != nil {
        log.Fatal(err)
    }
}
