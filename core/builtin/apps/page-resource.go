package apps

import (
	"bytes"
	"fmt"
	"lowcode/core/control"
	"lowcode/core/htmlrender"
	"path"
	"path/filepath"
	"time"
)

type StaticResource struct {
	Markdown htmlrender.UiResource `json:"markdown"`
	Html     htmlrender.UiResource `json:"html"`
}

func (s *StaticResource) Run(args []string, envs []string) (chan []byte, error) {

	fmt.Println("runcmdargs:", args)
	ch := make(chan []byte, 1)
	go func() {

		s.Markdown.Search(args[0], 1)

		var bts bytes.Buffer
		
		bts.WriteString(`{ "detas": [ { "header": "Today" },{"thumbnail": "https://cdn.vuetifyjs.com/images/lists/1.jpg","title": "Brunch this weekend?","subtitle": "<span>Ali Connors</span> Ill be in your neighborhood doing errands this weekend. Do you want to hang out?"}, { "divider": true, "inset": true },{"avatar": "https://cdn.vuetifyjs.com/images/lists/1.jpg","title": "Brunch this weekend?","subtitle": "<span>Ali Connors</span> Ill be in your neighborhood doing errands this weekend. Do you want to hang out?"},{ "header": "Updated" },{"avatar": "https://cdn.vuetifyjs.com/images/lists/1.jpg","title": "Brunch this weekend?","subtitle": "<span>Ali Connors</span> Ill be in your neighborhood doing errands this weekend. Do you want to hang out?"}] } `)
		ch <- bts.Bytes()
		time.Sleep(time.Second)
		close(ch)
	}()

	return ch, nil

	//return nil, fmt.Errorf("err")
}

func (s *StaticResource) inputControl(c *control.Control) (*control.IOControl, error) {
	in := &control.IOControl{Type: "form"}
	api := fmt.Sprintf("/%s/api", c.GetAppIndexPageHomePath())
	in.Stdin = append(in.Stdin, control.ParamMaps{Viewparam: "search", Paramtype: "txt", Required: true})
	in.View = control.ViewControl{Target: api, Type: "vue", Inline_string: fmt.Sprintf("<template> <xds-searchbox title='%s' posturl='%s' hint='%s' initFetchArgument='initFetchArgument' v-slot=\"{results}\" ><xdr-lists :items=\"results\"></xdr-lists></xds-searchbox> </template> <script>export default {name: 'resourceSearch', methods:{}, components:['xdsSearchbox', 'xdrLists']}</script>", c.Head.Title, api, "input search text")}

	return in, nil
}

func (s *StaticResource) outputControl() (*control.IOControl, error) {
	//in := &IOControl{Type: "json"}
	//in.Stdin = append(in.Stdin, ParamMaps{Viewparam: "search", Paramtype: "txt", Required: true})
	//in.View = ViewControl{Type: "vue", Filename: "/components/vue/vuetify/x-lists.vue" }
	//return in, nil
	return &control.IOControl{Type: "json"}, nil
}

func (s *StaticResource) entrypoint() (*control.ExeEntrypoint, error) {
	entry := &control.ExeEntrypoint{}
	entry.SetAppService(s)
	return entry, nil
}

func BuildStaticResource(c *control.Control) error {
	if !c.Resource.Markdown.Validate() && !c.Resource.Html.Validate() {
		return nil
	}

	Resource := &StaticResource{Markdown: c.Resource.Markdown, Html: c.Resource.Html}

	if in, err := Resource.inputControl(c); in != nil && err == nil {
		c.Input = *in
	} else {
		return err
	}

	if len(c.Output.Type) == 0 {
		if out, err := Resource.outputControl(); out != nil && err == nil {
			c.Output = *out
		} else {
			return err
		}

	}

	if en, err := Resource.entrypoint(); en != nil && err == nil {
		c.Entrypoint = *en
	} else {
		return err
	}

	//if have not vuetify
	c.Head.Links = append(c.Head.Links, "/assets/thirdparties/vuetify/fontfamily.css")
	c.Head.Links = append(c.Head.Links, "/assets/thirdparties/vuetify/materialdesignicons.min.css")
	c.Head.Links = append(c.Head.Links, "/assets/thirdparties/vuetify/vuetify.min.css")
	c.Head.Scripts = append(c.Head.Scripts, "/assets/thirdparties/vuetify/vue.js")
	c.Head.Scripts = append(c.Head.Scripts, "/assets/thirdparties/vuetify/vuetify.js")
	c.Head.Scripts = append(c.Head.Scripts, "/assets/js/base/core_iobind.js")

	savepath := path.Join(c.GetAppIndexPageHomePath(), "static-html")
	baseurl := fmt.Sprintf("/%s", savepath)
	if c.Resource.Markdown.Validate() {
		if err := c.Resource.Markdown.Init("markdown", c.GetAppIndexPageHomePath(), savepath, baseurl); err != nil {
			return err
		}

		if c.Resource.Markdown.Path[0] != '/' {
			c.Resource.Markdown.Path = filepath.Join(c.ControlFilePath, c.Resource.Markdown.Path)
		}

		if err := c.Resource.Markdown.Build(); err != nil {
			return err
		}
	}
	if c.Resource.Html.Validate() {
		if err := c.Resource.Html.Init("html", c.GetAppIndexPageHomePath(), savepath, baseurl); err != nil {
			return err
		}
		if c.Resource.Html.Path[0] != '/' {
			c.Resource.Html.Path = filepath.Join(c.ControlFilePath, c.Resource.Html.Path)
		}
		if err := c.Resource.Html.Build(); err != nil {
			return err
		}
	}

	return nil
}
