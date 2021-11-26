package main

import (
    "fmt"
    "log"
    "errors"
    "strings"
    "io/ioutil"
    "github.com/PuerkitoBio/goquery"
)

func numLeadingSpaces(line string) int {
    return len(line) - len(strings.TrimLeft(line, " "))
}

// Deindent takes a multiline string and returns a modified version with the
// same amount of leading spaces removed from all lines such that the first
// non-empty line has zero indentation.
func Deindent(content string) (string, error) {
    content = strings.Trim(content, "\n")
    lines := strings.Split(content, "\n")
    nSpacesFirst := numLeadingSpaces(lines[0])
    for i, line := range lines {
        if line == "" {
            continue
        }
        nSpaces := numLeadingSpaces(line)
        if nSpaces < nSpacesFirst {
            return "", errors.New(fmt.Sprintf(
                "Unexpected indentation: >>%s<<", line,
            ))
        }
        lines[i] = line[nSpacesFirst:]
    }
    return strings.Join(lines, "\n"), nil
}

// TempSourceFile returns the path to a temporary file with a unique name and
// the requested extension, populated with the requested content. Contents are
// passed through Deindent() before writing to file.
// It crashes if it hits any errors.
//
// It is the responsibility of the caller to delete the file after use.
func TempSourceFile(extension string, content string) string {
    // ioutil replaces the * in the pattern by a unique int at runtime
    fnamePattern := fmt.Sprintf("codex-temp*.%s", extension)
    // empty tmpdir means use system default
    tmpfile, err := ioutil.TempFile("", fnamePattern)
    if err != nil {
        log.Fatal(err)
    }

    deindented, err := Deindent(content)
    if err != nil {
        log.Fatal(err)
    }

    if _, err := tmpfile.Write([]byte(deindented)); err != nil {
        log.Fatal(err)
    }
    if err := tmpfile.Close(); err != nil {
        log.Fatal(err)
    }
    return tmpfile.Name()
}

func selText(sel *goquery.Selection) string {
    return strings.Trim(sel.Text(), " ")
}

func selCount(sel *goquery.Selection, selector string) int {
    return sel.Find(selector).Length()
}
