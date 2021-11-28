package main

import (
    "fmt"
    "log"
    "strings"
    "crypto/md5"
    "encoding/hex"
    "github.com/PuerkitoBio/goquery"
)

// PreNode is that which will turn into a node. It contains
type PreNode struct {
    // a single element Selection heading the node, eg an <h1>
    head *goquery.Selection

    // Each node has a body consisting of all subsequent siblings of the node's
    // head until and excluding the subsequent node head, eg everything
    // between an <h1> and the subsequent <h1>.
    body *goquery.Selection

    // Each node has a 0 based depth. Depths are relative to context,
    // independent of tag name. For example, an h3 can head a node of either
    // depths 0, 1, or 2.
    depth int
}

const tmpHeadClass string = "tmp-codex-head-class"

// Treeify is the main entrypoint for tree munging code. Its core
// traversal+transformation algo is implemented in treeify().
func Treeify(doc *goquery.Document) {
    doc.Find(strings.Join(HeadSelectors, ", ")).AddClass(tmpHeadClass)
    treeify(doc.Find("body").First(), 0)
    doc.Find("." + tmpHeadClass).RemoveClass(tmpHeadClass)
}

// Heads are elements in the DOM that trigger node creation. They become the
// node-head, and all their following siblings until the next head form the
// node-body.
var HeadSelectors = []string{"h1", "h2", "h3", "h4", "h5", "h6", "hr"}

// The rank of a heading is its index in HeadSelectors. The relative value of
// ranks between different nodes is what dictatese their relative tree position.
func rankOfHead(head *goquery.Selection) int {
    for i := 0; i < len(HeadSelectors); i++ {
        if head.Is(HeadSelectors[i]) {
            return i
        }
    }
    log.Fatal("unexpected head", SelectionToHtml(head))
    return -1
}

// nodify turns a PreNode into a Node, in place. This is a unit
// transformation of the DOM towards the new node tree structure.
//
// Before:
//      <h1> Title </h1>        <!-- prenode.head -->
//      <h2> Section </h2>      <!-- prenode
//      <p> Hello World </p>                .body -->
// After:
//      <node depth=0>
//        <node-head> <h1> Title </h1> </node-head>
//        <node-body>
//          <h2> Section </h2>
//          <p> Hello World </p>
//        </node-body>
//      </node>
func nodify(prenode PreNode) {
    var node *goquery.Selection

    if prenode.body.Length() == 0 {
        prenode.head.AfterHtml("<div> </div>")
        prenode.body = prenode.head.Next()
    }

    prenode.head.WrapAllHtml("<div class='node-head'> </div>")

    prenode.body.WrapAllHtml("<div class='node'> <div class='node-body'> </div> </div>")
    node = prenode.body.Parent().Parent()

    node.PrependSelection(prenode.head.Parent())
    node.SetAttr("class", fmt.Sprintf("node node-depth-%d", prenode.depth))
    node.SetAttr("id", fmt.Sprintf("node-%s", contentHash(node)))
}


// treeify recursively traverses the DOM and performs a sequence of in-place
// transformations that make the tree structure of the DOM match the semantic
// hierarchy of document sections, aka nodes.
func treeify(root *goquery.Selection, depth int) {
    if root.Length() > 1 {
        log.Fatal("expected a single root element!")
    }

    firstHead := root.ChildrenFiltered("." + tmpHeadClass).First()
    if firstHead.Length() == 0 {
        treeifyWithoutHeads(root, depth)
        return
    }

    leading := root.Children().Slice(0, firstHead.Index())
    if leading.Length() > 0 {
        leading.WrapAllHtml("<div>")
        treeifyWithoutHeads(leading.Parent(), depth)
    }

    curHead := firstHead
    var curBody, nextHead *goquery.Selection
    for curHead.Length() > 0 {
        // Given curHead H, nextHead is the first next sibling of H which is a
        // head with rank <= rank(H). All the nodes in between form the body of
        // the node rooted at H.
        nextHead = findNextHead(curHead)
        curBody = curHead.NextUntilSelection(nextHead)
        nodify(PreNode{curHead, curBody, depth})
        treeify(curBody.Parent(), depth + 1) // <= recurse

        curHead = nextHead
    }
}

func treeifyWithoutHeads(root *goquery.Selection, depth int) {
    root.Children().Each(func(i int, child *goquery.Selection) {
        child.BeforeHtml("<hr>")
        nodify(PreNode{
            head: child.Prev(),
            body: child,
            depth: depth,
        })
    })
}

func findNextHead(curHead *goquery.Selection) *goquery.Selection {
    curRank := rankOfHead(curHead)
    return curHead.NextAllFiltered(
        strings.Join(HeadSelectors[:curRank + 1], ", "),
    ).First()
}

func contentHash(node *goquery.Selection) string {
    html, err := goquery.OuterHtml(node)
    if err != nil {
        log.Fatal(err)
    }
    hash := md5.Sum([]byte(html))
    contentId := hex.EncodeToString(hash[:])[:8]
    return string(contentId)
}
