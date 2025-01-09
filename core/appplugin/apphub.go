package appplugin

import (
    "sync"
    "bytes"
    "strings"
    "github.com/blevesearch/bleve/v2"
	"onlinetools/core/control"
    "path/filepath"
)

var _defaultAppHub *AppHub

func DefaultAppHub() *AppHub {
    if _defaultAppHub == nil {
        _defaultAppHub = &AppHub{}
    }

    return _defaultAppHub
}


type AppInfo struct {
    Id string `json:"id"`
    Url string `json:"url"`
    Title string `json:"title"`
    Summary string `json:"summary"`
    Detail string `json:"detail"`
}

func (a *AppInfo) indexString() string {
    var bts bytes.Buffer
    bts.WriteString(a.Title)
    bts.WriteString(a.Summary)
    bts.WriteString(a.Detail)
    
    return bts.String()
}

type AppHub struct {
    apps []*AppInfo

    index bleve.Index
    sync.RWMutex
}

func (a *AppHub) Analysis(ctl *control.Control)(id, url, title, summary, detail string) {
    id = ctl.Name
    url = ctl.IndexPageUrl
    title = ctl.Head.Title
    summary = ctl.Head.Summary + strings.Join(ctl.Head.Keywords, ",")
    detail = ctl.Head.Description

    if ctl.Name == control.RootAppName {
        id, url = "", ""
    }

    return
}


func (a *AppHub) Append(id, url, title, summary, detail string) bool {
    if len(id) == 0 || len(url) == 0 || len(title) == 0 {
        return false
    }

    defer a.Unlock()
    a.Lock()

    has := false
    for _, in :=range a.apps {
        if in.Id == id {
            has = true
            break
        }
    }

    if !has {
        has = true
        a.apps = append(a.apps, &AppInfo{ Id:id, Url: url, Title: title, Summary: summary, Detail: detail })
        a.createIndex(a.apps[len(a.apps)-1])
    }

    return has
}

func (a *AppHub) Update(id, url, title, summary, detail string) {
    if len(id) == 0 || len(url) == 0 || len(title) == 0 {
        return 
    }

    a.Delete(id)

    defer a.Unlock()
    a.Lock()
    a.apps = append(a.apps, &AppInfo{ Id:id, Url: url, Title: title, Summary: summary, Detail: detail })
    a.createIndex(a.apps[len(a.apps)-1])
}

func (a *AppHub) Delete(id string) {
    defer a.Unlock()
    a.Lock()
    for i, in :=range a.apps {
        if in.Id == id {
            a.deleteIndex(a.apps[i])
            a.apps = append(a.apps[:i], a.apps[i+1:]...)
            break
        }
    }
}

func (a *AppHub) createIndex(info *AppInfo) {
	if a.index == nil {
		file := filepath.Join("", "._apps_.bleve")
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

	a.index.Index(info.Id , info.indexString())
}

func (a *AppHub) deleteIndex(info *AppInfo) {
    if a.index == nil || info == nil {
        return
    }
    a.index.Delete(info.Id)
}


func (a *AppHub) SearchByText(text string, maxcnts int) []*AppInfo {
    defer a.RUnlock()
    a.RLock()
	query := bleve.NewQueryStringQuery(text)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = maxcnts

	searchResult, _ := a.index.Search(searchRequest)

    var appInfos []*AppInfo
    for _, hit :=range searchResult.Hits {
        appInfos = append(appInfos, a.SearchById(hit.ID))
    }

    return appInfos
}

func (a *AppHub) SearchById(id string) *AppInfo {
    defer a.RUnlock()
    a.RLock()
    for i, in :=range a.apps {
        if in.Id == id {
            return a.apps[i]
        }
    }

    return nil
}

func (a *AppHub) GetRecentUpdated(count int) []*AppInfo {
    defer a.RUnlock()
    a.RLock()
    var appinfos []*AppInfo

    if count <= 0 {
        count = len(a.apps)
    }

    for l := len(a.apps)-1; l >= 0 && count > 0; {
        appinfos = append(appinfos, a.apps[l])
        l = l - 1
        count = count - 1
    }

    return appinfos
}
