package main

import _ "embed"

//go:embed static/codex.css
var codexCss string

//go:embed static/codex.js
var codexJs string

//go:embed static/pandoc.css
var pandocCss string

//go:embed static/codex.svg
var codexSvg string

type StaticFile struct {
	Body        string
	ContentType string
}

var STATICS = map[string]StaticFile{
	"/static/codex.css":  StaticFile{Body: codexCss, ContentType: "text/css"},
	"/static/codex.js":   StaticFile{Body: codexJs, ContentType: "text/javascript"},
	"/static/pandoc.css": StaticFile{Body: pandocCss, ContentType: "text/css"},
	"/static/codex.svg":  StaticFile{Body: codexSvg, ContentType: "image/svg+xml"},
}
