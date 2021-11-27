package main

import (
    "os"
    "testing"
    "github.com/stretchr/testify/assert"
)

func Test_H1_H1(t *testing.T) {
    //  root
    //  /  \
    // H1a  H1b
    fname := TempSourceFile("md", `
        # H1a

        H1a contents

        # H1b

        H1b contents`)
    doc, _ := render([]string{fname})
    defer os.Remove(fname)

    assert.Equal(t, 2, doc.Find(".node").Length())
    assert.Equal(t, 2, doc.Find(".node-depth-0").Length())
    assert.Equal(t, 0, doc.Find(".node-depth-1").Length())

    assert.Equal(t, 2, doc.Find(".node-head").Length())
    assert.Equal(t, 2, doc.Find(".node-body").Length())

    heads := doc.Find(".node-head")
    assert.Equal(t, "H1a", selText(heads.First()))
    assert.Equal(t, "H1b", selText(heads.Last()))

    bodies := doc.Find(".node-body")
    assert.Equal(t, "H1a contents", selText(bodies.First()))
    assert.Equal(t, "H1b contents", selText(bodies.Last()))
}


func Test_H1_H2(t *testing.T) {
    // H1
    //  |
    // H2
    fname := TempSourceFile("md", `
        # H1

        H1 contents

        ## H2

        H2 contents`)
    doc, _ := render([]string{fname})
    defer os.Remove(fname)

    assert.Equal(t, 2, doc.Find(".node").Length())
    assert.Equal(t, 1, doc.Find(".node-depth-0").Length())
    assert.Equal(t, 1, doc.Find(".node-depth-1").Length())
    assert.Equal(t, 0, doc.Find(".node-depth-2").Length())

    assert.Equal(t, "H1", selText(doc.Find(".node-depth-0 > .node-head")))
    assert.Equal(t, "H2", selText(doc.Find(".node-depth-1 > .node-head")))
    assert.Equal(t, "H2 contents", selText(doc.Find(".node-depth-1 .node-body")))

    assert.Equal(t, 1, doc.Find(".node-depth-0 .node-depth-1").Length())
}


func Test_H1_H3(t *testing.T) {
    // H1
    //  |
    // H3
    fname := TempSourceFile("md", `
        # H1

        H1 contents

        ### H3

        H3 contents`)
    doc, _ := render([]string{fname})
    defer os.Remove(fname)

    assert.Equal(t, 2, doc.Find(".node").Length())
    assert.Equal(t, 1, doc.Find(".node-depth-0").Length())
    assert.Equal(t, 1, doc.Find(".node-depth-1").Length())
    assert.Equal(t, 0, doc.Find(".node-depth-2").Length())

    assert.Equal(t, "H1", selText(doc.Find(".node-depth-0 > .node-head")))
    assert.Equal(t, "H3", selText(doc.Find(".node-depth-1 > .node-head")))
    assert.Equal(t, "H3 contents", selText(doc.Find(".node-depth-1 .node-body")))

    assert.Equal(t, 1, doc.Find(".node-depth-0 .node-depth-1").Length())
}


func Test_H2_H3_H6(t *testing.T) {
    // H2
    //  |
    // H3
    //  |
    // H6
    fname := TempSourceFile("md", `
        ## H2

        H2 contents

        ### H3

        H3 contents

        ##### H6

        H6 contents`)
    doc, _ := render([]string{fname})
    defer os.Remove(fname)
    // expect: H2 > H3 > H6

    assert.Equal(t, 3, doc.Find(".node").Length())
    assert.Equal(t, 1, doc.Find(".node-depth-0").Length())
    assert.Equal(t, 1, doc.Find(".node-depth-1").Length())
    assert.Equal(t, 1, doc.Find(".node-depth-2").Length())
    assert.Equal(t, 0, doc.Find(".node-depth-3").Length())

    assert.Equal(t, "H2", selText(doc.Find(".node-depth-0 > .node-head")))
    assert.Equal(t, "H3", selText(doc.Find(".node-depth-1 > .node-head")))
    assert.Equal(t, "H6", selText(doc.Find(".node-depth-2 > .node-head")))

    assert.Equal(t, 1, doc.Find(".node-depth-0 .node-depth-1").Length())
    assert.Equal(t, 1, doc.Find(".node-depth-0 .node-depth-2").Length())
    assert.Equal(t, 1, doc.Find(".node-depth-1 .node-depth-2").Length())
}


func Test_H1_H3_H2(t *testing.T) {
    //    H1
    //  /    \
    // H3    H2
    fname := TempSourceFile("md", `
        # H1

        H1 contents

        ### H3

        H3 contents

        ## H2

        H2 contents`)
    doc, _ := render([]string{fname})
    defer os.Remove(fname)

    assert.Equal(t, 3, doc.Find(".node").Length())
    assert.Equal(t, 1, doc.Find(".node-depth-0").Length())
    assert.Equal(t, 2, doc.Find(".node-depth-1").Length())
    assert.Equal(t, 0, doc.Find(".node-depth-2").Length())

    assert.Equal(t, "H1", selText(doc.Find(".node-depth-0 > .node-head")))
    assert.Equal(t, "H3", selText(doc.Find(".node-depth-1 > .node-head").First()))
    assert.Equal(t, "H2", selText(doc.Find(".node-depth-1 > .node-head").Last()))

    assert.Equal(t, 2, doc.Find(".node-depth-0").Find(".node-depth-1").Length())
}
