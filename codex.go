package main

import (
    "fmt"
    "log"
    "os"
    "errors"
    "os/exec"
    "syscall"
    "time"
    "sync"
    "bytes"
    "crypto/md5"
    "strings"
    "encoding/hex"
    "github.com/PuerkitoBio/goquery"
    "github.com/yosssi/gohtml"
)


var HeadSelectors = []string{"h1", "h2", "h3", "h4", "h5", "h6", "li:not(li li)"}

func SelectorRanks(elems []string) map[string]int {
    ranks := make(map[string]int)
    for idx, elem := range elems {
        ranks[elem] = idx
    }
    return ranks
}

func RankOfHead(head *goquery.Selection) int {
    for i := 0; i < len(HeadSelectors); i++ {
        if head.Is(HeadSelectors[i]) {
            fmt.Println("rank:", head.Text(), i, head.Length())
            return i
        }
    }
    log.Fatal("unexpected head!")
    return 0
}


const OutputTemplatePath = "static/index.html" // FIXME relpath, os.Executable()


func Mtime(path string) (time.Time, error) {
    fileinfo, err := os.Stat(path)
    if err != nil {
        return time.Time{}, err
    }
    stat := fileinfo.Sys().(*syscall.Stat_t)
    mtime := time.Unix(stat.Mtim.Sec, stat.Mtim.Nsec)
    return mtime, nil
}

// FIXME this is pandoc + Unflatten, all the heavy lifting. Massively simplify
func ConvertToHtmlDoc(path string, doc *goquery.Document, wg *sync.WaitGroup) error {
    defer wg.Done()
    mtime, err := Mtime(path)
    if err != nil {
        return err
    }

    cmd := exec.Command("pandoc", "-t", "html", path)

    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return err
    }

    err = cmd.Start()
    if err != nil {
        return err
    }

    doc_, err := goquery.NewDocumentFromReader(stdout)
    if (err != nil) {
        return err
    }
    *doc = *doc_

    err = cmd.Wait()
    if err != nil {
        return err
    }

    doc.Find(strings.Join(HeadSelectors, ", ")).AddClass("codex-head-cand") // FIXME name, document reason: h1, h2, h3 ... does not return children in DOM order
    Unflatten(doc.Find("body").First(), 0, 0)
    doc.Find(".codex-head-cand").RemoveClass("codex-head-cand")

    doc.Find(".node").Each(func(i int, sel *goquery.Selection) {
        sel.SetAttr("codex-source", path)
        // render mtime in ISO 8601 (RFC 3339), compatible with JS Date().
        sel.SetAttr("codex-mtime", mtime.Format(time.RFC3339))
    })

    return nil
}


// TODO error handling
func Unflatten(root *goquery.Selection, minHeadRank int, depth int) {
    if minHeadRank >= len(HeadSelectors) {
        return
    }

    firstHead := root.ChildrenFiltered(".codex-head-cand").First()
    if firstHead.Length() == 0 {
        return
    }
    firstRank := RankOfHead(firstHead)
    heads := root.ChildrenFiltered(strings.Join(HeadSelectors[:firstRank + 1], ", "))

    // TODO dom children before the first head should get their own node?
    heads.Each(func(i int, head *goquery.Selection) {
        var body, node *goquery.Selection
        if i + 1 >= heads.Length() {
            body = head.NextUntil("")
        } else {
            body = head.NextUntilNodes(heads.Get(i + 1))
        }

        if body.Length() == 0 {
            head.AfterHtml("<div> </div>")
            body = head.Next()
        }

        if head.Is("li") {
            node = head
            node.SetAttr("class", fmt.Sprintf("node li-node node-depth-%d", depth))
        } else {
            head.WrapAllHtml("<div class='node-head'> </div>")

            body.WrapAllHtml("<div class='node'> <div class='node-body'> </div> </div>")
            node = body.Parent().Parent()

            node.PrependSelection(head.Parent())
            node.SetAttr("class", fmt.Sprintf("node node-depth-%d", depth))

            Unflatten(body.Parent(), firstRank + 1, depth + 1) // <= recurse
        }

        node.SetAttr("id", fmt.Sprintf("node-%s", ContentHash(node)))
    })
}


func ContentHash(node *goquery.Selection) string {
    html, err := goquery.OuterHtml(node)
    if err != nil {
        log.Fatal(err)
    }
    hash := md5.Sum([]byte(html))
    contentId := hex.EncodeToString(hash[:])[:8]
    return string(contentId)
}

// FIXME refactor io.Reader
func LoadAsHtmlDoc(path string) (*goquery.Document, error) {
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


func BuildOutput(tplPath string, docs []goquery.Document) (*goquery.Document, error) {
    outDoc, err := LoadAsHtmlDoc(tplPath)
    if err != nil {
        return nil, err
    }

    var buffer bytes.Buffer

    for _, doc := range docs {
        html, err := doc.Find("body").First().Html()
        if err != nil {
            return nil, err
        }
        buffer.WriteString(html)
    }
    outDoc.Find("main").First().SetHtml(buffer.String())
    return outDoc, nil
}

// TODO error handling, see errgroup
func ConvertAll(paths []string) []goquery.Document {
    docs := make([]goquery.Document, len(paths))
    var wg sync.WaitGroup
    wg.Add(len(paths))
    for idx, path := range paths {
        // TODO errors?
        go ConvertToHtmlDoc(path, &docs[idx], &wg)
    }
    wg.Wait()
    return docs
}

func DocToHtml(doc *goquery.Document) (string, error) {
    html, err := doc.Html()
    if err != nil {
        return "", err
    }
    return gohtml.Format(html), nil
}

func render(paths []string) (*goquery.Document, error) {
    if len(paths) == 0 {
        return nil, errors.New("Need at least one input")
    }
    docs := ConvertAll(paths)
    out, err := BuildOutput(OutputTemplatePath, docs)
    if err != nil {
        return nil, err
        log.Fatal(err)
    }
    return out, nil
}

func main() {
    // TODO argparse
    inputs := os.Args[1:]
    out, err := render(inputs)
    html, err := DocToHtml(out)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(html)
}
