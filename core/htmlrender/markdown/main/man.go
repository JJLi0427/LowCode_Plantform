package main

import (
	"flag"
	"onlinetools/core/htmlrender/markdown"
)

func main() {

	f := flag.String("markdown", "", "file")
	flag.Parse()

	markdown.RebuildMarkdown2LocalHtml(*f, "")
}
