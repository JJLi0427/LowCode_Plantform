package textsearch

import (
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
)

type BleveUtils struct {
	IndexFileName string
	index         bleve.Index
}

func (a *BleveUtils) init() {
	if a.index == nil {
		if len(a.IndexFileName) == 0 {
			panic("BleveUtils: IndexFileName is empty")
		}
		d, f := filepath.Split(a.IndexFileName)
		dir := filepath.Join(TextSearchSavePathOfApps, d)
		os.MkdirAll(dir, 0700)
		file := filepath.Join(dir, f)

		if idx, err := bleve.Open(file); err != nil {
			mapping := bleve.NewIndexMapping()
			if idx, err := bleve.New(file, mapping); err != nil {
				panic(err)
			} else {
				a.index = idx
			}
		} else {
			a.index = idx
		}
	}
}

func (a *BleveUtils) UpdateIndex(id, text string) {
	a.init()
	a.index.Delete(id)
	a.index.Index(id, text)
}

func (a *BleveUtils) CreateIndex(id, text string) {
	a.init()

	if doc, err := a.index.Document(id); doc == nil || err != nil {
		a.index.Index(id, text)
	}

	//fmt.Println(id, text)
}

func (a *BleveUtils) DeleteIndex(id string) {
	a.init()
	if a.index != nil && len(id) > 0 {
		a.index.Delete(id)
	}
}

// return id slice
func (a *BleveUtils) SearchByText(text string, maxcnts int) []string {
	a.init()

	query := bleve.NewQueryStringQuery(text)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = maxcnts

	searchResult, _ := a.index.Search(searchRequest)

	var res []string
	for _, hit := range searchResult.Hits {
		res = append(res, hit.ID)
	}

	return res
}
