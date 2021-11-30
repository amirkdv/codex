package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"strings"
)

func numLeadingSpaces(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

/* Unindent takes a multiline string and uniformly unindents all lines such that
 * the first non-empty line has zero indentation. For example:
 *
 *    |\n
 *    |     Hello
 *    |       World
 *
 * becomes
 *
 *    |\n
 *    |Hello
 *    |  World
 */
func Unindent(content string) (string, error) {
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
// the requested extension and contents, crashes if it hits any errors.
//
// Contents are passed through Unindent() before writing to file.
//
// It is the responsibility of caller to delete the file after use.
func TempSourceFile(extension string, content string) string {
	// ioutil replaces the * in the pattern by a unique int at runtime
	fnamePattern := fmt.Sprintf("codex-temp-*.%s", extension)
	// empty tmpdir means use system default
	tmpfile, err := ioutil.TempFile("", fnamePattern)
	if err != nil {
		log.Fatal(err)
	}

	unindented, err := Unindent(strings.Trim(content, "\n"))
	if err != nil {
		log.Fatal(err)
	}

	if _, err := tmpfile.Write([]byte(unindented)); err != nil {
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
