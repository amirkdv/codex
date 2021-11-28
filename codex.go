package main

import (
    "io"
    "fmt"
    "log"
    "os"
    "errors"
    "os/exec"
    "syscall"
    "time"
    "bytes"
    "golang.org/x/sync/errgroup"
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


func ConvertToHtmlDoc(path string) (*goquery.Document, error) {
    mtime, err := Mtime(path)
    if err != nil {
        return nil, err
    }

    cmd := exec.Command("pandoc", "-t", "html", path)

    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return nil, err
    }
    stderr, err := cmd.StderrPipe()
    if err != nil {
        return nil, err
    }

    if err = cmd.Start(); err != nil {
        return nil, err
    }

    doc, err := goquery.NewDocumentFromReader(stdout)
    if err != nil {
        return nil, err
    }

    stderrContents, err := io.ReadAll(stderr)
    if err != nil {
        return nil, err
    }

    if err = cmd.Wait(); err != nil {
        return nil, errors.New(string(stderrContents))
    }

    Treeify(doc)

    doc.Find(".node").Each(func(i int, sel *goquery.Selection) {
        sel.SetAttr("codex-source", path)
        // render mtime in ISO 8601 (RFC 3339), compatible with JS Date().
        sel.SetAttr("codex-mtime", mtime.Format(time.RFC3339))
    })

    return doc, nil
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
    // HACK outDoc.Find(".node:not(:has(.node))").AddClass("codex-leaf")
    return outDoc, nil
}

func ConvertAll(paths []string) ([]goquery.Document, error) {
    docs := make([]goquery.Document, len(paths))

    var errg errgroup.Group
    for idx, path := range paths {
        path := path
        idx := idx
        errg.Go(func() error {
            doc, err := ConvertToHtmlDoc(path)
            if err != nil {
                return err
            }
            docs[idx] = *doc
            return nil
        })
    }
    if err := errg.Wait(); err != nil {
        return nil, err
    }
    return docs, nil
}

func render(paths []string) (*goquery.Document, error) {
    if len(paths) == 0 {
        return nil, errors.New("Need at least one input")
    }
    docs, err := ConvertAll(paths)
    if err != nil {
        log.Fatal(err)
    }
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
