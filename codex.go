package main

import (
    "fmt"
    "log"
    "os"
    "os/exec"

    "github.com/PuerkitoBio/goquery"
)


// equivalent of transform() in node impl
func toHtml(input string, output string) {
    css := "style.css"
    header := "header.html"
    title := "J"
    pandocArgs := []string{
        "-s",  "-t", "html",
        "--metadata", fmt.Sprintf("title=%s", title),
        input, "-o", output}

    // TODO capture stderr instead of mysterious "exit status 1", eg when
    // header.html is missing
    _, err := exec.Command("pandoc", pandocArgs...).Output()
    if err != nil {
        log.Print("error!")
        log.Fatal(err)
    }
}

func main() {
    input := "example.md"
    output := "output.html"

    toHtml(input, output)
}
