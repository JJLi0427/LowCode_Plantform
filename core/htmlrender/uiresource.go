package htmlrender

import (
	"fmt"
	"onlinetools/core/common/file"
	"onlinetools/core/htmlrender/markdown"
	"onlinetools/core/sitemap"
	"os"
	"path/filepath"
	"time"

	"github.com/blevesearch/bleve/v2"
)

type MatchedResource struct {
	Avater    string `json:"avater,omitempty"`
	Thumbnail string `json:"thumbnail,omitempty"`
	Image     string `json:"image,omitempty"`

	Title    string `json:"title,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`
	Summary  string `json:"summary,omitempty"`
	Detail   string `json:"detail,omitempty"`
	Text     string `json:"text,omitempty"`

	Link   string `json:"link,omitempty"`
	Target string `json:"target,omitempty"`

	Rate int `json:"rate,omitempty"`
}

type UiResource struct {
	Path   string `json:"path"`
	Render string `json:"render"`
	Layout string `json:"layout"`

	resourceType string
	basepath     string
	savepath     string
	baseurl      string
	index        bleve.Index
}

func (u *UiResource) Validate() bool {
	return len(u.Path) > 0
}

func (u *UiResource) Init(resourceType, basepath, savepath, baseurl string) error {
	u.resourceType = resourceType
	u.basepath = basepath
	u.savepath = savepath
	u.baseurl = baseurl
	if err := os.MkdirAll(u.savepath, 0750); err != nil {
		return err
	}
	return nil
}

func (u *UiResource) Build() error {
	fch, err := u.ScanPath()
	if err != nil {
		return err
	}
	if _, err := u.RenderHtml(fch); err != nil {
		return err
	}

	return nil
}

func (u *UiResource) ScanPath() (chan string, error) {
	chfile := make(chan string, 10)
	inittm := time.Date(1990, time.February, 1, 23, 0, 0, 0, time.UTC)
	scanner := func(path, ext string) {
		for {
			nfiles, _ := file.ReadFilesInDir(path, true, func(f os.FileInfo, d string) bool {
				if ff := filepath.Ext(f.Name()); ff == ext {
					if inittm.Before(f.ModTime()) {
						return true
					}
				}
				return false
			})

			inittm = time.Now()

			for _, f := range nfiles {
				chfile <- f
			}

			time.Sleep(time.Minute * 10)
		}
	}

	if u.resourceType == "md" || u.resourceType == "markdown" {
		go scanner(u.Path, ".md")
	} else {
		return nil, fmt.Errorf("not supported resourcetype currently")
	}

	return chfile, nil
}

func (u *UiResource) createIndex(md, url string) {
	if u.index == nil {
		file := filepath.Join(u.basepath, "index.bleve")
		if idx, err := bleve.Open(file); err != nil {
			mapping := bleve.NewIndexMapping()
			if idx, err := bleve.New(file, mapping); err != nil {
				panic(err)
			} else {
				u.index = idx
			}
		} else {
			u.index = idx
		}

	}

	if bts, err := os.ReadFile(md); err == nil {
		u.index.Index(url, string(bts))
	}

}

func (u *UiResource) RenderHtml(chSrc chan string) (chan string, error) {
	mdrender := func() {
		for f := range chSrc {
			if rurl, err := markdown.RebuildMarkdown2LocalHtml(f, u.savepath); err == nil && len(rurl) > 0 {
				sitemap.GetDefaultSitemapBuilder().AddSiteMapofUrl(u.baseurl + "/" + rurl)
				u.createIndex(f, u.baseurl+"/"+rurl)
			}
		}

		u.index.Close()
		u.index = nil
	}

	if u.resourceType == "md" || u.resourceType == "markdown" {
		go mdrender()
	} else {
		return nil, fmt.Errorf("not supported resourcetype currently")
	}

	return nil, nil
}

func (u *UiResource) Search(text string, maxcnts int) []*MatchedResource {
	if u.index == nil {
		return nil
	}

	query := bleve.NewQueryStringQuery(text)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = maxcnts

	searchResult, _ := u.index.Search(searchRequest)

	fmt.Println(searchResult.String())

	return nil
}
