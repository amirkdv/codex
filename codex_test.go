package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_H1_p(t *testing.T) {
	// H1
	//  |
	//  p
	fname := TempSourceFile("md", `
        # H1

        Hello World
        `)
	doc, _ := Codex([]string{fname})
	defer os.Remove(fname)

	assert.Equal(t, 2, doc.Find(".node").Length())
	assert.Equal(t, 1, doc.Find(".node-depth-0").Length())
	assert.Equal(t, 1, doc.Find(".node-depth-1").Length())

	assert.Equal(t, "H1", selText(doc.Find(".node-depth-0 .node-head")))
	assert.Equal(t, "Hello World", selText(doc.Find(".node-depth-1 .node-body").Last()))
}

func Test_H2_p_p(t *testing.T) {
	// gist:
	//  - H2 still gets depth 0 because codex depths are relative
	//  - When there are no heads left all children become a node of their own
	//   H2
	//  /  \
	// p    p
	fname := TempSourceFile("md", `
        ## H2

        Hello World

        Goodbye World
        `)
	doc, _ := Codex([]string{fname})
	defer os.Remove(fname)

	assert.Equal(t, 3, doc.Find(".node").Length())
	assert.Equal(t, 1, doc.Find(".node-depth-0").Length())
	assert.Equal(t, 2, doc.Find(".node-depth-1").Length())

	assert.Equal(t, "H2", selText(doc.Find(".node-depth-0 .node-head")))
	assert.Equal(t, "Hello World", selText(doc.Find(".node-depth-1 .node-body").First()))
	assert.Equal(t, "Goodbye World", selText(doc.Find(".node-depth-1 .node-body").Last()))
}

func Test_H1_H1(t *testing.T) {
	//  root
	//  /  \
	// H1a  H1b
	fname := TempSourceFile("md", `
        # H1a

        # H1b

        `)
	doc, _ := Codex([]string{fname})
	defer os.Remove(fname)

	assert.Equal(t, 2, doc.Find(".node").Length())
	assert.Equal(t, 2, doc.Find(".node-depth-0").Length())

	assert.Equal(t, "H1a", selText(doc.Find(".node-head").First()))
	assert.Equal(t, "H1b", selText(doc.Find(".node-head").Last()))

	assert.Equal(t, 2, doc.Find(".node-body").Length())
}

func Test_H1_p_H1(t *testing.T) {
	//  root
	//  /  \
	// H1a  H1b
	//  |
	//  p
	fname := TempSourceFile("md", `
        # H1a

        Hello World

        # H1b
        `)
	doc, _ := Codex([]string{fname})
	defer os.Remove(fname)

	assert.Equal(t, 3, doc.Find(".node").Length())
	assert.Equal(t, 2, doc.Find(".node-depth-0").Length())
	assert.Equal(t, 1, doc.Find(".node-depth-1").Length())

	assert.Equal(t, "H1a", selText(doc.Find(".node-head").First()))
	assert.Equal(t, "H1b", selText(doc.Find(".node-head").Last()))

	paragraph := doc.Find(".node-depth-1 .node-body").First()
	assert.Equal(t, "Hello World", selText(paragraph))
}

func Test_H1_p_p_H2(t *testing.T) {
	//     H1
	//   / | \
	//  p  p  H2
	fname := TempSourceFile("md", `
        # H1

        Hello World

        Goodbye World

        ## H2
        `)
	doc, _ := Codex([]string{fname})
	defer os.Remove(fname)

	assert.Equal(t, 4, doc.Find(".node").Length())
	assert.Equal(t, 1, doc.Find(".node-depth-0").Length())
	assert.Equal(t, 3, doc.Find(".node-depth-1").Length())

	paragraphs := doc.Find(".node-depth-1 .node-body p")
	assert.Equal(t, "Hello World", selText(paragraphs.First()))
	assert.Equal(t, "Goodbye World", selText(paragraphs.Last()))
}

func Test_H1_H2(t *testing.T) {
	// H1
	//  |
	// H2
	fname := TempSourceFile("md", `
        # H1

        ## H2

        `)
	doc, _ := Codex([]string{fname})
	defer os.Remove(fname)

	assert.Equal(t, 2, doc.Find(".node").Length())
	assert.Equal(t, 1, doc.Find(".node-depth-0").Length())
	assert.Equal(t, 1, doc.Find(".node-depth-1").Length())

	assert.Equal(t, "H1", selText(doc.Find(".node-depth-0 > .node-head")))
	assert.Equal(t, "H2", selText(doc.Find(".node-depth-1 > .node-head")))

	assert.Equal(t, 1, doc.Find(".node-depth-0 .node-depth-1").Length())
}

func Test_H1_p_H2_p(t *testing.T) {
	//   H1
	//  /  \
	// p    H2
	//       \
	//        p
	fname := TempSourceFile("md", `
        # H1

        Hello World

        ## H2

        Goodbye World`)
	doc, _ := Codex([]string{fname})
	defer os.Remove(fname)

	assert.Equal(t, 4, doc.Find(".node").Length())
	assert.Equal(t, 1, doc.Find(".node-depth-0").Length())
	assert.Equal(t, 2, doc.Find(".node-depth-1").Length())
	assert.Equal(t, 1, doc.Find(".node-depth-2").Length())

	assert.Equal(t, "H1", selText(doc.Find(".node-depth-0 > .node-head")))
	assert.Equal(t, "H2", selText(doc.Find(".node-depth-1 > .node-head").Last()))

	hello := doc.Find(".node-depth-1 .node-body").First()
	assert.Equal(t, "Hello World", selText(hello))

	goodbye := doc.Find(".node-depth-1 .node-body").Last()
	assert.Equal(t, "Goodbye World", selText(goodbye))
}

func Test_H1_H3(t *testing.T) {
	// gist: should be the same as H1 > H2 because codex depths are relative.
	// H1
	//  |
	// H3
	fname := TempSourceFile("md", `
        # H1

        ### H3
        `)
	doc, _ := Codex([]string{fname})
	defer os.Remove(fname)

	assert.Equal(t, 2, doc.Find(".node").Length())
	assert.Equal(t, 1, doc.Find(".node-depth-0").Length())
	assert.Equal(t, 1, doc.Find(".node-depth-1").Length())

	assert.Equal(t, "H1", selText(doc.Find(".node-depth-0 > .node-head")))
	assert.Equal(t, "H3", selText(doc.Find(".node-depth-1 > .node-head")))

	assert.Equal(t, 1, doc.Find(".node-depth-0 .node-depth-1").Length())
}

func Test_H2_H3_H6(t *testing.T) {
	// gist: should be the same as H1 > H2 > H3 because codex depths are relative.
	// H2
	//  |
	// H3
	//  |
	// H6
	fname := TempSourceFile("md", `
        ## H2

        ### H3

        ##### H6
        `)
	doc, _ := Codex([]string{fname})
	defer os.Remove(fname)
	// expect: H2 > H3 > H6

	assert.Equal(t, 3, doc.Find(".node").Length())
	assert.Equal(t, 1, doc.Find(".node-depth-0").Length())
	assert.Equal(t, 1, doc.Find(".node-depth-1").Length())
	assert.Equal(t, 1, doc.Find(".node-depth-2").Length())

	assert.Equal(t, "H2", selText(doc.Find(".node-depth-0 > .node-head")))
	assert.Equal(t, "H3", selText(doc.Find(".node-depth-1 > .node-head")))
	assert.Equal(t, "H6", selText(doc.Find(".node-depth-2 > .node-head")))

	assert.Equal(t, 1, doc.Find(".node-depth-0 .node-depth-1").Length())
	assert.Equal(t, 1, doc.Find(".node-depth-0 .node-depth-2").Length())
	assert.Equal(t, 1, doc.Find(".node-depth-1 .node-depth-2").Length())
}

func Test_H1_H3_H2_H3(t *testing.T) {
	// note: this is important! Getting this scenario right is why the
	// node building algo has to traverse the DOM one head at a time.
	//    H1
	//   /  \
	// H3    H2
	//        \
	//         H3
	fname := TempSourceFile("md", `
        # H1

        ### H3 depth 1

        ## H2

        ### H3 depth 2
        `)
	doc, _ := Codex([]string{fname})
	defer os.Remove(fname)

	assert.Equal(t, 4, doc.Find(".node").Length())
	assert.Equal(t, 1, doc.Find(".node-depth-0").Length())
	assert.Equal(t, 2, doc.Find(".node-depth-1").Length())
	assert.Equal(t, 1, doc.Find(".node-depth-2").Length())

	assert.Equal(t, "H1", selText(doc.Find(".node-depth-0 > .node-head")))
	assert.Equal(t, "H3 depth 1", selText(doc.Find(".node-depth-1 > .node-head").First()))
	assert.Equal(t, "H2", selText(doc.Find(".node-depth-1 > .node-head").Last()))
	assert.Equal(t, "H3 depth 2", selText(doc.Find(".node-depth-2 > .node-head")))
}

func Test_md_rst(t *testing.T) {
	fnameMd := TempSourceFile("md", `
        # MD

        Hello World MD
        `)
	fnameRst := TempSourceFile("rst", `
        RST
        ===

        Hello World RST
        `)
	doc, _ := Codex([]string{fnameMd, fnameRst})
	defer os.Remove(fnameMd)
	defer os.Remove(fnameRst)

	assert.Equal(t, 4, doc.Find(".node").Length())
	assert.Equal(t, 2, doc.Find(".node-depth-0").Length())
	assert.Equal(t, 2, doc.Find(".node-depth-1").Length())

	assert.Equal(t, "MD", selText(doc.Find(".node-depth-0 > .node-head").First()))
	assert.Equal(t, "RST", selText(doc.Find(".node-depth-0 > .node-head").Last()))

	assert.Equal(t, "Hello World MD", selText(doc.Find(".node-depth-1 .node-body").First()))
	assert.Equal(t, "Hello World RST", selText(doc.Find(".node-depth-1 .node-body").Last()))
}
