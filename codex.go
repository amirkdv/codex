package main

import (
	"bytes"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"
)

const (
	parallelism = 2 // TODO CLI arg
)

// A Codex document (a codoc) corresponds to a path containing some sort of
// markup/down, any format supported by pandoc. Formats are infered from file
// extension by pandoc, with markdown as fallback default.
type Document struct {
	Path string
}

// Codex holds the context for a single instance of the codex app.
// It's intended to be instantiated only once, using CLI arguments.
type Codex struct {
	Inputs         []Document
	buildSemaphore chan int
	outputDoc      *goquery.Document
	outputHtml     string
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
	codocs := make([]Document, len(paths))
	for idx, path_ := range paths {
		codocs[idx] = Document{Path: path_}
	}
	cdx := Codex{
		Inputs:         codocs,
		buildSemaphore: make(chan int, parallelism),
	}
	return &cdx, nil
}

func (codoc Document) Mtime() time.Time {
	fileinfo, err := os.Stat(codoc.Path)
	if err != nil {
		log.Fatal(err)
	}
	stat := fileinfo.Sys().(*syscall.Stat_t)
	mtime := time.Unix(stat.Mtim.Sec, stat.Mtim.Nsec)
	return mtime
}

func (cdx *Codex) TransformAll() ([]*goquery.Document, error) {
	htmlDocs := make([]*goquery.Document, len(cdx.Inputs))

	var errg errgroup.Group
	for idx, codoc := range cdx.Inputs {
		// because closure around goroutine below
		idx := idx
		codoc := codoc
		errg.Go(func() error {
			doc, err := cdx.Transform(codoc)
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

func (cdx *Codex) Convert(codoc Document) (*goquery.Document, error) {
	cdx.buildSemaphore <- 1 // up the semaphore, blocks until channel has space
	defer func() {
		log.Println("<<", codoc.Path)
		<-cdx.buildSemaphore // down the semaphore
	}()

	log.Println(">>", codoc.Path)
	cmd := exec.Command("pandoc", "-t", "html", codoc.Path)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	htmlDoc, err := goquery.NewDocumentFromReader(stdout)
	if err != nil {
		return nil, err
	}

	errMsg, err := io.ReadAll(stderr)
	if err != nil {
		return nil, err
	}

	if err = cmd.Wait(); err != nil {
		return nil, errors.New(string(errMsg))
	}

	return htmlDoc, nil
}

func (cdx *Codex) Transform(codoc Document) (*goquery.Document, error) {
	htmlDoc, err := cdx.Convert(codoc)
	if err != nil {
		return nil, err
	}

	Treeify(htmlDoc)

	htmlDoc.Find(".node").Each(func(i int, sel *goquery.Selection) {
		sel.SetAttr("codex-source", codoc.Path)
		// render mtime in ISO 8601 (RFC 3339), compatible with JS Date().
		sel.SetAttr("codex-mtime", codoc.Mtime().Format(time.RFC3339))
	})

	return htmlDoc, nil
}

func (cdx *Codex) Output() string {
	return cdx.outputHtml
}
