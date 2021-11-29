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
	"path"
)

func StaticDir() string {
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	return path.Join(path.Dir(exe), "static")
}

var OutputTemplatePath = path.Join(StaticDir(), "index.html")

type Codex struct {
	inputs []*Codocument
}

func NewCodex(paths []string) (*Codex, error) {
	if len(paths) == 0 {
		return nil, errors.New("Need at least one input")
	}
	codocs := make([]*Codocument, len(paths))
	for idx, path := range paths {
		codoc, err := NewCodocument(path)
		if err != nil {
			log.Fatal(err)
		}
		codocs[idx] = codoc
	}
	return &Codex{codocs}, nil
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

func (cdx Codex) Build() (*goquery.Document, error) {
	docs, err := cdx.TransformAll()
	if err != nil {
		return nil, err
	}

	outDoc, err := LoadHtml(CodexOutputTemplate)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer

	for _, doc := range docs {
		html, err := doc.Find("body").First().Html()
		if err != nil {
			return nil, err
		}
		buffer.WriteString(html)
	}
	outDoc.Find("main").First().SetHtml(buffer.String())
	outDoc.Find(".node:not(:has(.node))").AddClass("node-leaf")

	log.Println("Finished building")
	return outDoc, nil
}

func (cdx Codex) BuildAndWatch() {
	out, err := cdx.Build()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(DocToHtml(out)) // FIXME serve

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	waitForever := make(chan bool)
	go cdx.buildOnWrite(watcher)

	for _, codoc := range cdx.inputs {
		if err = watcher.Add(codoc.path); err != nil {
			log.Fatal(err)
		}
	}
	<-waitForever
}

func (cdx Codex) buildOnWrite(watcher *fsnotify.Watcher) {
	log.Println("Watching", len(cdx.inputs), "docs for changes ...")
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println(event.Name, "has changed: rebuilding ...")
				out, _ := cdx.Build() // FIXME ws send
				fmt.Println(DocToHtml(out))
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}
