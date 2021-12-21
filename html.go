package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/yosssi/gohtml"
	"log"
	"strings"
)

func InnerHtml(sel *goquery.Selection) string {
	html, err := sel.Html()
	if err != nil {
		log.Fatal(err)
	}
	return html
}

func OuterHtml(sel *goquery.Selection) string {
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

func LoadHtml(html string) (*goquery.Document, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}
	return doc, nil
}
