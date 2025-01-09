package sitemap

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

var defaultSitemapBuilder *SiteMapBuilder

func GetDefaultSitemapBuilder() *SiteMapBuilder {
	return defaultSitemapBuilder
}

func init() {
	defaultSitemapBuilder = &SiteMapBuilder{}
}

type siteRecord struct {
	Url    string `xml:"loc"`
	Mod        string `xml:"lastmod,omitempty"`
	Changefreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

func (s *siteRecord) Init(url, mod, freq, pri string) *siteRecord {
	s.Url = url
	s.Mod = mod         //2005-05-10T17:33:30+08:00
	s.Changefreq = freq //always,hourly,daily,weekly,monthly,yearly,never
	s.Priority = pri    //0.0 - 1.0
	return s
}
func (s *siteRecord) Size() int {
	return len(s.Url) + len(s.Mod) + len(s.Changefreq) + len(s.Priority) + 150
}

type siteMap struct {
	XMLName     xml.Name      `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	Sites       []*siteRecord `xml:"url"`
	size, count int
}

func (s *siteMap) AddRecord(url string) {
	site := (&siteRecord{}).Init(url, time.Now().Format("2006-01-02"), "daily", "0.8")
	s.Sites = append(s.Sites, site)

	s.size += site.Size()
	s.count++
}

func (s *siteMap) Size() int {
	return s.size
}

func (s *siteMap) Count() int {
	return s.count
}

func (s *siteMap) Bytes() ([]byte, error) {

	return xml.MarshalIndent(s, " ", " ")
}

func (s *siteMap) Clear() {
	s.size = 0
	s.count = 0
	s.Sites = s.Sites[0:0]
}

type sitemapIndex struct {
	XMLName xml.Name `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 sitemapindex"`
	ASites  []struct {
		Url string `xml:"loc"`
	} `xml:"sitemap"`
}

type SiteMapBuilder struct {
	sync.Mutex
	SiteBaseUrl       string
	SavePath          string
	MaxCountsEveryMap int
	MaxSizeEveryMap   int
	siteMap
	siteMapCounts  int
	sitemapUpdated bool
	sitemapIndex
}

func (s *SiteMapBuilder) Init(siteBaseUrl, savePath string, maxCountsEveryMap, maxSizeEveryMap int) *SiteMapBuilder {
	defer s.Unlock()
	s.Lock()
	s.SiteBaseUrl = siteBaseUrl
	s.SavePath = savePath
	s.MaxCountsEveryMap = maxCountsEveryMap
	s.MaxSizeEveryMap = maxSizeEveryMap

	return s
}

func (s *SiteMapBuilder) newSiteMapName(siteMapBaseUrl string, idx int) string {
	name := fmt.Sprintf("sitemap-%d.xml", idx)

	if l := len(s.ASites); l == 0 || s.ASites[l-1].Url != siteMapBaseUrl+name {
		s.ASites = append(s.ASites, struct {
			Url string "xml:\"loc\""
		}{siteMapBaseUrl + name})
		s.sitemapUpdated = false
	}

	return name
}

func (s *SiteMapBuilder) siteMaps() ([]byte, error) {
	return xml.MarshalIndent(s.sitemapIndex, " ", " ")
}

func (s *SiteMapBuilder) BuildSiteMaps() {
	if s.sitemapUpdated {
		return
	}

	if bts, err := s.siteMaps(); err == nil {
		ioutil.WriteFile(path.Join(s.SavePath, "sitemap.xml"), bts, os.ModePerm)
		s.sitemapUpdated = true
	} else {
		fmt.Println(err)
	}
}

func (s *SiteMapBuilder) BuildOnce(idx int) {
	sitemapurl := s.SiteBaseUrl + "/sitemap/"

	if len(s.SavePath) > 0 {
		if fs, err := os.Stat(s.SavePath); err != nil {
			os.MkdirAll(s.SavePath, os.ModePerm)
		} else if !fs.IsDir() {
			panic("savepath for sitMap is not directory: " + s.SavePath)
		}
	}

	if bts, err := s.Bytes(); err == nil {
		ioutil.WriteFile(path.Join(s.SavePath, s.newSiteMapName(sitemapurl, idx)), bts, os.ModePerm)
		s.BuildSiteMaps()
	}
}

func (s *SiteMapBuilder) AddSiteMapofUrl(url string) {
	s.Lock()
	defer s.Unlock()
	if len(url) == 0 {
		return
	}

	if !strings.HasPrefix(url, "http") && url[0] == '/' {
		url = fmt.Sprintf("%s%s", s.SiteBaseUrl, url)
	}

	if s.MaxCountsEveryMap <= 0 {
		s.MaxCountsEveryMap = 50000
	}

	if s.MaxSizeEveryMap <= 0 {
		s.MaxSizeEveryMap = 10 * 1024 * 1024 //10MB
	}

	if len(s.SavePath) <= 0 {
		s.SavePath = "views/sitemap/"
	}

	s.AddRecord(url)

	if s.Count() >= s.MaxCountsEveryMap || s.Size()+1024 >= s.MaxSizeEveryMap {
		s.BuildOnce(s.siteMapCounts)
		s.siteMapCounts += 1
		s.Clear()
	} else {
		s.BuildOnce(s.siteMapCounts)
	}
}
