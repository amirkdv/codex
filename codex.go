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
    "github.com/PuerkitoBio/goquery"
)


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

    if err = cmd.Start(); err != nil {
        return err
    }

    doc_, err := goquery.NewDocumentFromReader(stdout)
    if err != nil {
        return err
    }
    *doc = *doc_

    err = cmd.Wait()
    if err != nil {
        return err
    }

    Treeify(doc)

    doc.Find(".node").Each(func(i int, sel *goquery.Selection) {
        sel.SetAttr("codex-source", path)
        // render mtime in ISO 8601 (RFC 3339), compatible with JS Date().
        sel.SetAttr("codex-mtime", mtime.Format(time.RFC3339))
    })

    return nil
}

func BuildOutput(tplPath string, docs []goquery.Document) (*goquery.Document, error) {
    outDoc, err := LoadHtmlPath(tplPath)
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
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(DocToHtml(out))
}
