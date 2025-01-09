package textsearch

import "fmt"

type LightSearch struct {
	IndexFileName string
}

func (l *LightSearch) UpdateIndex(id, text string) {

}
func (l *LightSearch) CreateIndex(id, text string) {
	fmt.Println(id, text, (&WordsSpliter{}).Split(text))

}
func (l *LightSearch) DeleteIndex(id string) {

}
func (l *LightSearch) SearchByText(text string, maxcnts int) []string {

	return nil
}
