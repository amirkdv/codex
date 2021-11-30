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
	for idx, path_ := range paths {
		codocs[idx] = &Codocument{path: path_}
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

func (cdx *Codex) Output() string {
	return DocToHtml(cdx.output)
}
