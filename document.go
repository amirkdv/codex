package main

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// Codocument is a Codex document. It corresponds to a path on disk containing
// some sort of markup/down, any format supported by pandoc. Formats are
// infered from file extension by pandoc.
type Codocument struct {
	path string
}

func (codoc Codocument) Mtime() time.Time {
	fileinfo, err := os.Stat(codoc.path)
	if err != nil {
		log.Fatal(err)
	}
	stat := fileinfo.Sys().(*syscall.Stat_t)
	mtime := time.Unix(stat.Mtim.Sec, stat.Mtim.Nsec)
	return mtime
}

func (codoc Codocument) ToHtml() (*goquery.Document, error) {
	cmd := exec.Command("pandoc", "-t", "html", codoc.path)

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

func (codoc Codocument) Transform() (*goquery.Document, error) {
	htmlDoc, err := codoc.ToHtml()
	if err != nil {
		return nil, err
	}

	Treeify(htmlDoc)

	htmlDoc.Find(".node").Each(func(i int, sel *goquery.Selection) {
		sel.SetAttr("codex-source", codoc.path)
		// render mtime in ISO 8601 (RFC 3339), compatible with JS Date().
		sel.SetAttr("codex-mtime", codoc.Mtime().Format(time.RFC3339))
	})

	return htmlDoc, nil
}
