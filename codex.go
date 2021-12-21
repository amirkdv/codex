package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
	"log"
)

const (
	PandocConcurrency = 3 // maximum number of pandoc subprocesses
)

// Codex holds the context for a single instance of the codex app.
// It's intended to be instantiated only once, using CLI arguments.
type Codex struct {
	Inputs  map[string]*Document
	HtmlDoc *goquery.Document
	HtmlStr string

	pandocPool *PandocPool
}

func NewCodex(paths []string) (*Codex, error) {
	if len(paths) == 0 {
		return nil, errors.New("Need at least one input")
	}

	codocs := make(map[string]*Document)
	for _, filePath := range paths {
		codocs[filePath] = NewDocument(filePath)
	}

	cdx := Codex{
		Inputs:     codocs,
		pandocPool: NewPandocPool(PandocConcurrency),
	}

	doc, err := cdx.DOMSkeleton()
	if err != nil {
		return nil, err
	}
	cdx.HtmlDoc = doc

	log.Println("Starting with", len(cdx.Inputs), "input document(s)")
	if err := cdx.BuildAll(); err != nil {
		return nil, err
	}
	log.Println("Finished building from", len(cdx.Inputs), "docs")

	return &cdx, nil
}

// DOMSkeleton loads the codex HTML template and creates stand-in
// <article> elements in <main> for each of the input Documents.
//    <html> ... <body>
//      <main>
//        <article codex-source="example.md" ...> </article>
//        <article codex-source="other.rst" ...> </article>
//        ...
//      </main>
//    </body> </html>
func (cdx *Codex) DOMSkeleton() (*goquery.Document, error) {
	doc, err := LoadHtml(CodexOutputTemplate)
	if err != nil {
		return nil, err
	}

	main := doc.Find("main")

	for _, codoc := range cdx.Inputs {
		main.AppendHtml(fmt.Sprintf(`<article codex-source="%s"/>`, codoc.Path))
	}
	return doc, nil
}

// CurrentDOMArticle returns a goquery Selection containing the current DOM
// <article> corresponding to the given input Document.
func (cdx *Codex) CurrentDOMArticle(codoc *Document) *goquery.Selection {
	selector := fmt.Sprintf(`article[codex-source="%s"]`, codoc.Path)
	article := cdx.HtmlDoc.Find(selector)
	if article.Length() == 0 {
		log.Fatal(errors.New(fmt.Sprintf("Unexpected input doc: %s", codoc.Path)))
	}
	return article
}

// Update rebuilds the specified document and updates its DOM <article>.
func (cdx *Codex) Update(codoc *Document) (string, error) {
	article := cdx.CurrentDOMArticle(codoc)
	innerHtml, err := cdx.Transform(codoc)
	if err != nil {
		return "", err
	}
	article.SetHtml(innerHtml)
	article.SetAttr("codex-mtime", ToIso8601(codoc.Mtime))
	cdx.HtmlStr = DocToHtml(cdx.HtmlDoc)
	return OuterHtml(article), nil
}

// Transform takes an input Document and returns it as codex HTML.
func (cdx *Codex) Transform(codoc *Document) (string, error) {
	codoc.CheckMtime()
	codoc.SetBtime()

	htmlDoc, err := cdx.pandocPool.Run(codoc.Path)
	if err != nil {
		return "", err
	}
	Treeify(htmlDoc)
	return InnerHtml(htmlDoc.Find("body")), nil
}

func (cdx *Codex) BuildAll() error {
	var errg errgroup.Group
	for _, codoc := range cdx.Inputs {
		codoc := codoc // because closure below
		errg.Go(func() error {
			_, err := cdx.Update(codoc)
			if err != nil {
				return err
			}
			return nil
		})
	}
	if err := errg.Wait(); err != nil {
		return err
	}

	cdx.HtmlStr = DocToHtml(cdx.HtmlDoc)
	return nil
}

func (cdx *Codex) Html() string {
	return cdx.HtmlStr
}
