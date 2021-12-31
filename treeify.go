package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"log"
	"strings"
)

// PreNode is that which will turn into a node.
type PreNode struct {
	// a single-element Selection heading the node, eg an <h1>
	Head *goquery.Selection

	// Node body consists of all subsequent siblings of the node's
	// head until and excluding the subsequent node head, eg everything
	// between an <h1> and the subsequent <h1>.
	Body *goquery.Selection

	// Each node has a 0 based depth. Depths are relative to context,
	// independent of tag name. For example, an h3 can head a node of either
	// depths 0, 1, or 2.
	Depth int
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
// ranks between different nodes is what dictates their relative tree position.
func rankOfHead(head *goquery.Selection) int {
	for i := 0; i < len(HeadSelectors); i++ {
		if head.Is(HeadSelectors[i]) {
			return i
		}
	}
	log.Fatal("unexpected head", OuterHtml(head))
	return -1
}

// nodify turns a PreNode into a Node, in place. This is a unit
// transformation of the DOM towards the new node tree structure.
//
// Before:
//      <h1> Title </h1>        <!-- prenode.Head -->
//      <h2> Section </h2>      <!-- prenode
//      <p> Hello World </p>                .Body -->
// After:
//      <node depth=0>
//        <node-head> <h1> Title </h1> </node-head>
//        <node-body>
//          <h2> Section </h2>
//          <p> Hello World </p>
//        </node-body>
//      </node>
func nodify(prenode PreNode) *goquery.Selection {
	var node *goquery.Selection

	if prenode.Body.Length() == 0 {
		prenode.Head.AfterHtml("<div> </div>")
		prenode.Body = prenode.Head.Next()
	}

	if prenode.Body.Is("ul") {
		prenode.Body.ChildrenFiltered("li").Each(func(i int, li *goquery.Selection) {
			nodifyListItem(li.Nodes[0])
			li.SetAttr("class", fmt.Sprintf("node node-depth-%d", prenode.Depth+1))
			li.SetAttr("id", fmt.Sprintf("node-%s", contentHash(li)))
		})
	}

	prenode.Head.WrapAllHtml("<div class='node-head'> </div>")

	prenode.Body.WrapAllHtml("<div class='node'> <div class='node-body'> </div> </div>")
	node = prenode.Body.Parent().Parent()

	node.PrependSelection(prenode.Head.Parent())
	node.SetAttr("class", fmt.Sprintf("node node-depth-%d", prenode.Depth))
	node.SetAttr("id", fmt.Sprintf("node-%s", contentHash(node)))
	return node
}

// nodifyListItem is a special case handler for lists
// An <li> is a special kind of node in the sense that:
//	1. its head is the first text node of the li, requires digging deeper from
//	   goquery tools to html.Node.
//	2. its head is not an immediate child of body, violating the big assumption.
func nodifyListItem(liNode *html.Node) {
	if liNode.FirstChild == nil {
		return
	}
	if liNode.FirstChild.Type != html.TextNode {
		// typically this happens for ElementType, cf stdlib's html/node.go
		// nodifying li's that start with elements, eg <p>, requires a refactor
		return
	}
	headSel := goquery.Selection{Nodes: []*html.Node{liNode.FirstChild}}
	headSel.WrapHtml("<span class='node-head'></span>")

	var bodyNodes []*html.Node
	curNode := liNode.FirstChild.NextSibling
	for curNode != nil {
		bodyNodes = append(bodyNodes, curNode)
		curNode = curNode.NextSibling
	}
	bodySel := goquery.Selection{Nodes: bodyNodes}
	bodySel.WrapHtml("<div class='node-body'></span>")
}

// treeify recursively traverses the DOM and performs a sequence of in-place
// transformations that make the tree structure of the DOM match the semantic
// hierarchy of document sections, aka nodes.
func treeify(root *goquery.Selection, depth int) {
	if root.Length() > 1 {
		log.Fatal("expected a single root element!")
	}

	// caution: the point of tmpHeadClass is the following query.
	// If we simply concatenate head selectors with ",", the resulting selection
	// will *not* necessarily be in correct tree order. For example if you ask
	// for `h1, h2, h3` in `<h3>...</h3> ... <h2>...</h2>` you'll get the h2
	// before the h3.
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
		treeify(curBody.Parent(), depth+1) // <= recurse

		curHead = nextHead
	}
}

func treeifyWithoutHeads(root *goquery.Selection, depth int) {
	root.Children().Each(func(i int, child *goquery.Selection) {
		child.BeforeHtml("<div/>")
		node := nodify(PreNode{
			Head:  child.Prev(),
			Body:  child,
			Depth: depth,
		})
		node.AddClass("headless")
	})
}

func findNextHead(curHead *goquery.Selection) *goquery.Selection {
	curRank := rankOfHead(curHead)
	return curHead.NextAllFiltered(
		strings.Join(HeadSelectors[:curRank+1], ", "),
	).First()
}

func contentHash(node *goquery.Selection) string {
	htmlStr, err := goquery.OuterHtml(node)
	if err != nil {
		log.Fatal(err)
	}
	hash := md5.Sum([]byte(htmlStr))
	contentId := hex.EncodeToString(hash[:])[:8]
	return string(contentId)
}
