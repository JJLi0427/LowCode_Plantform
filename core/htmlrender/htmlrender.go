package htmlrender

type HtmlRender interface {
	ScanPath() (chan string, error)
	RenderHtml(chan string) (chan string, error)
	Search(text string, count int)
}
