package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/yosssi/gohtml"
	"log"
	"os"
)

func SelectionToHtml(sel *goquery.Selection) string {
	html, err := goquery.OuterHtml(sel)
	if err != nil {
		log.Fatal(err)
	}
	return html
}

func DocToHtml(doc *goquery.Document) string {
	html, err := doc.Html()
	if err != nil {
		log.Fatal(err)
	}
	return gohtml.Format(html)
}

func LoadHtmlPath(path string) (*goquery.Document, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
