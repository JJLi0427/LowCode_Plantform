// Package parser provides methods to grab all links from markdown files.
package markdown

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"lowcode/core/htmlrender/downloader"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

func dirname(fname string) (string, string) {

	sub := strings.TrimSuffix(fname, filepath.Ext(fname))

	return url.PathEscape(sub), sub
}

func RebuildMarkdown2LocalHtml(md, savepath string) (string, error) {

	if ext := filepath.Ext(md); ext != ".md" {

		return "", fmt.Errorf("check file ext failed, markdown parse")
	}
	relativeurl, sub := dirname(filepath.Base(md))
	savepath = path.Join(savepath, sub)
	os.MkdirAll(savepath, 0700)

	markdown, err := os.ReadFile(md)
	if err != nil {
		return "", err
	}

	var newmarkdown bytes.Buffer
	re := regexp.MustCompile(`(?m)(^!\[\]\(([^)]+)\))?`)

	scanner := bufio.NewScanner(bytes.NewReader(markdown))
	// Scans line by line
	for scanner.Scan() {
		// Make regex
		matches := re.FindAllStringSubmatch(scanner.Text(), -1)

		if len(matches) == 1 && len(matches[0]) == 3 && strings.HasPrefix(matches[0][2], "http") {
			if fname, err := downloader.DownloadResource(matches[0][2], savepath); err == nil {
				newmarkdown.WriteString(strings.ReplaceAll(scanner.Text(), matches[0][2], fname))
				newmarkdown.WriteByte('\n')
			} else {
				newmarkdown.WriteString(scanner.Text())
				newmarkdown.WriteByte('\n')
			}

		} else {
			newmarkdown.WriteString(scanner.Text())
			newmarkdown.WriteByte('\n')
		}

		if newmarkdown.Len() > 0 {
			if bts, err := Markdown2Html(newmarkdown.Bytes(), true); err == nil {
				os.WriteFile(path.Join(savepath, "index.html"), bts, 0700)
			}
		}
	}

	return relativeurl, nil
}

// ParseLink parses a line and grabs the Link, Title and the Description attached to it.
// The format of the line should be as follows: `- [Title](Link) - Description`.
// Description can be omitted.
func ParseLink(line string) map[string]string {
	// Holds all the title, link, and description
	m := make(map[string]string)

	// Regex to extract title, link, and description
	re := regexp.MustCompile(`(?m)(^- \[([^]]+)\]\(([^)]+)\) ?-? ?(.*)?)?`)

	// Make regex
	match := re.FindStringSubmatch(line)

	m["Title"] = ""
	m["Link"] = ""
	m["Description"] = ""
	if len(match) == 5 {
		m["Title"] = match[2]
		m["Link"] = match[3]
		m["Description"] = match[4]
	}

	return m
}

// ParseImageLink parses an image line and grabs the Link attached to it.
// The format of the line should be as follows: `![](Link)`.
func ParseImageLink(line string) string {
	// Regex to extract title, link, and description
	re := regexp.MustCompile(`(?m)(^!\[\]\(([^)]+)\))?`)

	// Make regex
	match := re.FindStringSubmatch(line)

	link := ""
	if len(match) == 3 {
		link = match[2]
	}

	return link
}

// GetAllLinks returns all links and their names from a given markdown file.
func GetAllLinks(markdown string) map[string]string {
	// Holds all the links and their corresponding values
	m := make(map[string]string)

	// Regex to extract link and text attached to link
	re := regexp.MustCompile(`\[([^\]]*)\]\(([^)]*)\)`)

	scanner := bufio.NewScanner(strings.NewReader(markdown))
	// Scans line by line
	for scanner.Scan() {
		// Make regex
		matches := re.FindAllStringSubmatch(scanner.Text(), -1)

		// Only apply regex if there are links and the link does not start with #
		if matches != nil {
			if strings.HasPrefix(matches[0][2], "#") == false {
				// fmt.Println(matches[0][2])
				m[matches[0][1]] = matches[0][2]
			}
		}
	}
	return m
}

// fileToString returns string representation of a file.
func fileToString(file string) (string, error) {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	s := string(bytes)
	return s, nil
}

// ParseMarkdownFile parses a markdown file and returns all markdown links from it.
func ParseMarkdownFile(fileName string) (map[string]string, error) {
	file, err := fileToString(fileName)
	if err != nil {
		log.Fatal()
	}
	return GetAllLinks(file), nil
}

// DownloadURL returns Body response from the URL.
func DownloadURL(URL string) (string, error) {
	resp, err := http.Get(URL)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	return buf.String(), nil
}

// ParseMarkdownURL parses an URL and returns all markdown links from it.
func ParseMarkdownURL(URL string) (map[string]string, error) {
	file, err := DownloadURL(URL)
	if err != nil {
		return make(map[string]string), err
	}
	return GetAllLinks(file), nil
}