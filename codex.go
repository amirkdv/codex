package main

import (
	"bytes"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"path"
)

// Codex holds the context for a single instance of the codex app.
// It's intended to be instantiated only once, using CLI arguments.
type Codex struct {
	Inputs		[]*Document
	outputDoc	*goquery.Document
	outputHtml	string
}

func RootDir() string {
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	return path.Dir(exe)
}

func NewCodex(paths []string) (*Codex, error) {
	if len(paths) == 0 {
		return nil, errors.New("Need at least one input")
	}
	codocs := make([]*Document, len(paths))
	for idx, path_ := range paths {
		codocs[idx] = &Document{Path: path_}
	}
	return &Codex{Inputs: codocs}, nil
}

func (cdx *Codex) TransformAll() ([]*goquery.Document, error) {
	htmlDocs := make([]*goquery.Document, len(cdx.Inputs))

	var errg errgroup.Group
	for idx, codoc := range cdx.Inputs {
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

	cdx.outputDoc = outDoc
	cdx.outputHtml = DocToHtml(cdx.outputDoc)
	return nil
}

func (cdx *Codex) Output() string {
	return cdx.outputHtml
}
