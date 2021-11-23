package main

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "syscall"
    "time"
    "bytes"
    "crypto/md5"
    "encoding/hex"
    "github.com/PuerkitoBio/goquery"
)

import "reflect"


var HeadSelectors = [...]string{"h1", "h2", "h3", "h4", "h5", "li:not(li li)"}
const OutputTemplatePath = "index.html"


func getMtime(path string) (time.Time, error) {
    fileinfo, err := os.Stat(path)
    if err != nil {
        return time.Time{}, err
    }
    stat := fileinfo.Sys().(*syscall.Stat_t)
    mtime := time.Unix(stat.Mtim.Sec, stat.Mtim.Nsec)
    return mtime, nil
}


func ConvertToHtmlDoc(path string) (*goquery.Document, error) {
    mtime, err := getMtime(path)
    if err != nil {
        return nil, err
    }

    cmd := exec.Command("pandoc", "-t", "html", path)

    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return nil, err
    }

    err = cmd.Start()
    if err != nil {
        return nil, err
    }

    doc, err := goquery.NewDocumentFromReader(stdout)
    if (err != nil) {
        return nil, err
    }

    err = cmd.Wait()
    if err != nil {
        return nil, err
    }


    Unflatten(doc.Find("body").First(), HeadSelectors[:])

    doc.Find(".node").Each(func(i int, sel *goquery.Selection) {

        sel.SetAttr("codex-source", path)
        sel.SetAttr("codex-mtime", mtime.Format(time.RFC3339)) // iso 8601, JS: new Date("2021-11-20T00:36:20-05:00").toLocaleString()
    })

    return doc, nil
}

func dir(value interface{}) {
    type_ := reflect.TypeOf(value)
    for i := 0; i < type_.NumMethod(); i++ {
        fmt.Println(type_.Method(i).Name)
    }
}


func Unflatten(root *goquery.Selection, selectors []string) {
    if len(selectors) == 0 {
        return
    }

    var heads *goquery.Selection
    idx := 0

    for idx < len(selectors) {
        heads = root.Find(selectors[idx])
        if heads.Length() > 0 {
            break
        }
        idx ++
    }

    if heads.Length() == 0 {
        return
    }

    heads.Each(func(i int, head *goquery.Selection) {
        var body, node *goquery.Selection
        if i + 1 >= heads.Length() {
            body = head.NextUntil("")
        } else {
            body = head.NextUntilNodes(heads.Get(i + 1))
        }

        depth := len(HeadSelectors) - len(selectors)
        if head.Is("li") {
            node = head
            node.SetAttr("class", fmt.Sprintf("node li-node node-depth-%d", depth))
        } else {
            head.WrapAllHtml("<div class='node-head'> </div>")

            body.WrapAllHtml("<div class='node'> <div class='node-body'> </div> </div>")
            node = body.Parent().Parent()

            node.PrependSelection(head.Parent())
            node.SetAttr("class", fmt.Sprintf("node node-depth-%d", depth))

            Unflatten(body.Parent(), selectors[idx + 1:]) // <= recurse
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
    return fmt.Sprintf("%s", contentId)
}


func LoadDoc(path string) (*goquery.Document, error) {
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


func BuildOutput(docs []*goquery.Document) (*goquery.Document, error) {
    outDoc, err := LoadDoc(OutputTemplatePath)
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

func render(paths []string) (string, error) {
    docs := make([]*goquery.Document, len(paths))
    for idx, path := range paths {
        doc, err := ConvertToHtmlDoc(path)
        if err != nil {
            return "", err
        }
        docs[idx] = doc
    }
    output, err := BuildOutput(docs)
    if err != nil {
        return "", err
    }

    html, err := output.Html()
    if err != nil {
        return "", err
    }
    return html, nil
}


func main() {
    inputs := os.Args[1:]
    if len(inputs) == 0 {
        log.Fatal("Expected at least one argument")
    }
    html, err := render(inputs)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(html)
}
