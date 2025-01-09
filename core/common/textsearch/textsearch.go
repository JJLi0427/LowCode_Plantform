package textsearch

const TextSearchSavePathOfApps = ".appSearchIndexs"

type engineType int

const (
	BelveSearch engineType = iota

	//small text speedup engine
	Light
)

type IndexTextSearch interface {
	UpdateIndex(id, text string)
	CreateIndex(id, text string)
	DeleteIndex(id string)

	//search by text, return associated text's ids
	SearchByText(text string, maxcnts int) []string
}

func NewIndexTextSearch(indexType engineType, indexName string) IndexTextSearch {
	var engine IndexTextSearch = nil

	switch indexType {
	case BelveSearch:
		engine = &BleveUtils{IndexFileName: indexName}
	case Light:
		engine = &LightSearch{IndexFileName: indexName}
	}

	return engine
}
