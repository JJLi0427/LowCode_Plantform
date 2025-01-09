package apps

import (
	"fmt"
	"encoding/json"
	"onlinetools/core/control"
	"onlinetools/core/appplugin"
	"onlinetools/core/htmlrender"
)


type homePageApi struct {
	apphub *appplugin.AppHub
}

func (h *homePageApi)Run(args []string, envs []string) (chan []byte, error) {
	if len(args) == 0 || h.apphub == nil {
		return nil, fmt.Errorf("need input argument, or init error")
	}

	ch := make(chan []byte, 1)

	if args[0] == "initFetchArgument" {
		ch <- h.convertMsg(h.apphub.GetRecentUpdated(30))
	}else{
		ch <- h.convertMsg(h.apphub.SearchByText(args[0], 8))
	}

	close(ch)

	return ch, nil
}

func (h *homePageApi)convertMsg(appInfos []*appplugin.AppInfo) []byte {
	if len(appInfos) == 0 {
		return []byte(`{"result":[]}`)
	}

	type Msg struct {
		Bts []*htmlrender.MatchedResource `json:"results"`
	}
	msg := &Msg{}

	for _, appinfo :=range appInfos {
		msg.Bts = append(msg.Bts, 
			&htmlrender.MatchedResource{
				Title: appinfo.Title,
				Summary: appinfo.Summary,
				Detail: appinfo.Detail,
				Link: appinfo.Url})
	}

	bts, _ := json.Marshal(msg)

	return bts
}


func BuildHomePage(c *control.Control) error {
	if c.Name != control.RootAppName  ||  c.Entrypoint.HasExeEntrypoint() {
		return nil
	}

	c.Entrypoint.SetAppService(&homePageApi{apphub: appplugin.DefaultAppHub()})

	return nil
}
