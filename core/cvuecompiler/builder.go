package cvuecompiler

import (
	"bytes"
	"fmt"
	"onlinetools/core/control"
	"os"
	"path/filepath"
	"strings"
)

type componentSet struct {
	components []*VueComponent
}

func (c *componentSet) Has(d string) bool {
	for _, s := range c.components {
		if s.Name == d {
			return true
		}
	}
	return false
}

func (c *componentSet) ToMap() map[string]*VueComponent {
	da := make(map[string]*VueComponent)

	for _, s := range c.components {
		da[s.Name] = s
	}
	return da
}

func (c *componentSet) Append(d *VueComponent) bool {
	if !c.Has(d.Name) {
		c.components = append(c.components, d)
		return true
	}

	return false
}

func (c *componentSet) GetSlice() []*VueComponent {
	return c.components
}

func (c *componentSet) Remove(name string) bool {
	for i, s := range c.components {
		if s.Name == name {
			c.components = append(c.components[:i], c.components[i+1:]...)
			return true
		}
	}

	return false
}

func ReadFilesInDir(dir string, withpath bool, accept func(os.FileInfo, string) bool) ([]string, error) {
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("VueBuilder.LoadAllComponents failed for dir[%s], %s", dir, err.Error())
	}

	var files []string
	if d, err := os.Open(dir); err == nil {
		defer d.Close()
		if infos, err := d.Readdir(-1); err == nil {
			for _, info := range infos {
				if info.IsDir() {
					if fs, err := ReadFilesInDir(dir+"/"+info.Name(), withpath, accept); err == nil {
						files = append(files, fs...)
					} else {
						return nil, err
					}
				} else if accept(info, dir) {
					if withpath {
						files = append(files, dir+"/"+info.Name())
					} else {
						files = append(files, info.Name())
					}
				}
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}

	return files, nil
}

type ResultOfVueBuilder struct {
	El  string
	Js  string
	Css string
}
type VueBuilder struct {
	Components []*VueComponent
}

func (v *VueBuilder) Generate(control *control.Control) error {
	if control.Input.View.Type != "vue" && control.Output.View.Type != "vue" {
		return nil
	}

	hasVuetify := false
	for _, s := range control.Head.Scripts {
		if strings.HasSuffix(strings.ToLower(s), "vuetify.js") {
			hasVuetify = true
			break
		}
	}

	if !hasVuetify {
		for _, s := range control.Tail.Scripts {
			if strings.HasSuffix(strings.ToLower(s), "vuetify.js") {
				hasVuetify = true
				break
			}
		}
	}

	if control.Input.View.Type == "vue" && control.Output.View.Type == "vue" {
		res, err := v.BuildIOVue(control, "vueoielem", hasVuetify)
		if err != nil {
			return err
		}
		control.Input.View.Inline_string = res.El
		control.Output.View.Inline_string = ""
		control.SaveJS(res.Js, false)
		control.SaveCSS(res.Css, true)
		control.Head.Scripts = append(control.Head.Scripts, "/assets/js/base/core_iobind.js")
	} else {
		var ignore map[string]*VueComponent
		if control.Input.View.Type == "vue" {
			res, err, childs := v.BuildJS(control.Input.View.Inline_string, control.Input.View.Filename, "vueinputelem", true, nil, hasVuetify)
			if err != nil {
				return err
			}
			ignore = childs
			control.Input.View.Inline_string = res.El
			control.SaveJS(res.Js, false)
			control.SaveCSS(res.Css, true)
		}
		if control.Output.View.Type == "vue" {
			res, err, _ := v.BuildJS(control.Output.View.Inline_string, control.Output.View.Filename, "vueoutputelem", false, ignore, hasVuetify)
			if err != nil {
				return err
			}
			control.Head.Scripts = append(control.Head.Scripts, "/assets/js/base/core_iobind.js")
			control.Output.View.Inline_string = res.El
			control.SaveJS(res.Js, false)
			control.SaveCSS(res.Css, true)
		}

	}

	return nil
}

func (v *VueBuilder) FindAllChilds(c *VueComponent, childs *componentSet) error {
	for _, cname := range c.ChildComponents {
		find := false
		for i, s := range v.Components {
			if s.Name == cname {
				find = true
				if childs.Append(v.Components[i]) {
					if err := v.FindAllChilds(v.Components[i], childs); err != nil {
						return err
					}
				}
			}
		}

		if !find {
			return fmt.Errorf("VueBuilder can't find %s component", cname)
		}
	}

	return nil
}

func (v *VueBuilder) BuildIOVue(control *control.Control, elementId string, vuetify bool) (*ResultOfVueBuilder, error) {
	var ivueEntry *VueComponent = &VueComponent{HasVuetify: vuetify}
	var ovueEntry *VueComponent = &VueComponent{HasVuetify: vuetify,
		AddAddtionalDivs: []string{`<div id='selfinteraldataentryid' style='display:hidden;' @selfinteraldataentry='selfinteraldataentrymethod($event)'></div>`},
		AddAddtionalMethods: []string{`
selfinteraldataentrymethod(e){
    if (! e?.detail) {
        return
    }
    for (let key in e.detail) {
        this.$data[key] = e.detail[key]
    }
}`}}

	ovueEntry.Scripts = `;vuedispatchDataConfig('#selfinteraldataentryid', 'selfinteraldataentry');`

	if err := ivueEntry.BuildFromComponentFileContent([]byte(control.Input.View.Inline_string)); err != nil {
		return nil, err
	}

	if len(ivueEntry.Name) == 0 && !ivueEntry.GenerateComponentNameByFileName(control.Input.View.Filename) {
		return nil, fmt.Errorf("GenerateComponentNameByFileName[%s] failed, should specify [name] attr in vue [export default]", control.Input.View.Filename)
	}

	if err := ovueEntry.BuildFromComponentFileContent([]byte(control.Output.View.Inline_string)); err != nil {
		return nil, err
	}

	if len(ovueEntry.Name) == 0 && !ovueEntry.GenerateComponentNameByFileName(control.Output.View.Filename) {
		return nil, fmt.Errorf("GenerateComponentNameByFileName[%s] failed, should specify [name] attr in vue [export default]", control.Output.View.Filename)
	}

	ivueEntry.Name = "i" + ivueEntry.Name
	ovueEntry.Name = "o" + ovueEntry.Name

	childs := &componentSet{}
	if len(ivueEntry.ChildComponents) != 0 {
		if len(v.Components) == 0 {
			return nil, fmt.Errorf("VueBuilder: should load all vue components before Build()")
		}
		if err := v.FindAllChilds(ivueEntry, childs); err != nil {
			return nil, err
		}
	}

	if len(ovueEntry.ChildComponents) != 0 {
		if len(v.Components) == 0 {
			return nil, fmt.Errorf("VueBuilder: should load all vue components before Build()")
		}
		if err := v.FindAllChilds(ivueEntry, childs); err != nil {
			return nil, err
		}
	}

	var bts bytes.Buffer
	var btt bytes.Buffer
	var bty bytes.Buffer

	bts.WriteString(ivueEntry.Scripts)
	bts.WriteByte('\n')
	bts.WriteString(ovueEntry.Scripts)
	bts.WriteByte('\n')

	bty.WriteString(ivueEntry.Styles)
	bty.WriteByte('\n')
	bty.WriteString(ovueEntry.Styles)
	bty.WriteByte('\n')

	for _, c := range childs.GetSlice() {
		if len(c.Scripts) > 0 {
			bts.WriteString(c.Scripts)
			bts.WriteByte('\n')
		}

		if len(c.Styles) > 0 {
			bty.WriteString(c.Styles)
			bty.WriteByte('\n')
		}
	}

	schilds := childs.GetSlice()
	for i := len(schilds) - 1; i >= 0; i-- {
		c := schilds[i]
		if c.Name == ivueEntry.Name || c.Name == ovueEntry.Name {
			continue
		}
		bts.WriteString("Vue.component('")
		bts.WriteString(c.Name)
		bts.WriteString("',")
		bts.WriteString(c.VueExportDefault)
		bts.WriteString(");\n")
	}
	bts.WriteString("Vue.component('")
	bts.WriteString(ivueEntry.Name)
	bts.WriteString("',")
	bts.WriteString(ivueEntry.VueExportDefault)
	bts.WriteString(");\n")
	bts.WriteString("Vue.component('")
	bts.WriteString(ovueEntry.Name)
	bts.WriteString("',")
	bts.WriteString(ovueEntry.VueExportDefault)
	bts.WriteString(");\n")

	//props:{resource:{type:Object, Require:true}}

	bts.WriteString("new Vue({el:'")
	bts.WriteString(fmt.Sprintf("div#%s.cc%s", elementId, ivueEntry.Name))
	bts.WriteString("',")
	if vuetify {
		bts.WriteString("vuetify: new Vuetify(),")
	}
	bts.WriteString("components:{")
	//bts.WriteString(entry.Name)
	bts.WriteString("},data:{},")
	bts.WriteString("})")

	btt.WriteString(fmt.Sprintf("<div id='%s' class='cc%s'>", elementId, ivueEntry.Name))

	if vuetify {
		btt.WriteString(fmt.Sprintf("<v-app><v-main><%s></%s><%s></%s></v-main></v-app>", ivueEntry.Name, ivueEntry.Name, ovueEntry.Name, ovueEntry.Name))
	} else {
		btt.WriteString(fmt.Sprintf("<%s></%s><%s></%s>", ivueEntry.Name, ivueEntry.Name, ovueEntry.Name, ovueEntry.Name))
	}
	btt.WriteString("</div>")

	R := &ResultOfVueBuilder{
		El:  btt.String(),
		Js:  bts.String(),
		Css: bty.String(),
	}
	return R, nil
}

func (v *VueBuilder) BuildJS(vuecontent, vuefilename string, elementId string, isinput bool, ignore map[string]*VueComponent, vuetify bool) (*ResultOfVueBuilder, error, map[string]*VueComponent) {
	var vueEntry *VueComponent = &VueComponent{HasVuetify: vuetify}

	if !isinput {
		vueEntry = &VueComponent{HasVuetify: vuetify,
			AddAddtionalDivs: []string{`<div id='selfinteraldataentryid' style='display:hidden;' @selfinteraldataentry='selfinteraldataentrymethod($event)'></div>`},
			AddAddtionalMethods: []string{`
selfinteraldataentrymethod(e){
    if (! e?.detail) {
        return
    }
    for (let key in e.detail) {
        this.$data[key] = e.detail[key]
    }
}`}}

		vueEntry.Scripts = `;vuedispatchDataConfig('#selfinteraldataentryid', 'selfinteraldataentry');`
	}

	if err := vueEntry.BuildFromComponentFileContent([]byte(vuecontent)); err != nil {
		return nil, err, nil
	}

	if len(vueEntry.Name) == 0 && !vueEntry.GenerateComponentNameByFileName(vuefilename) {
		return nil, fmt.Errorf("GenerateComponentNameByFileName[%s] failed, should specify [name] attr in vue [export default]", vuefilename), nil
	}

	if isinput {
		vueEntry.Name = "i" + vueEntry.Name
	} else {
		vueEntry.Name = "o" + vueEntry.Name
	}

	childs := &componentSet{}
	if len(vueEntry.ChildComponents) != 0 {
		if len(v.Components) == 0 {
			return nil, fmt.Errorf("VueBuilder: should load all vue components before Build()"), nil
		}
		if err := v.FindAllChilds(vueEntry, childs); err != nil {
			return nil, err, nil
		}
	}

	if ignore != nil {
		for k, _ := range ignore {
			childs.Remove(k)
		}
	}

	a, b := v.build(vueEntry, childs, elementId)
	return a, b, childs.ToMap()
}

func (v *VueBuilder) build(entry *VueComponent, childs *componentSet, elmounted string) (*ResultOfVueBuilder, error) {
	var bts bytes.Buffer
	var btt bytes.Buffer
	var bty bytes.Buffer

	if len(entry.Scripts) > 0 {
		bts.WriteString(entry.Scripts)
		bts.WriteByte('\n')
	}
	if len(entry.Styles) > 0 {
		bty.WriteString(entry.Styles)
		bty.WriteByte('\n')
	}
	for _, c := range childs.GetSlice() {
		if len(c.Scripts) > 0 {
			bts.WriteString(c.Scripts)
			bts.WriteByte('\n')
		}
		if len(c.Styles) > 0 {
			bty.WriteString(c.Styles)
			bty.WriteByte('\n')
		}
	}

	schilds := childs.GetSlice()
	for i := len(schilds) - 1; i >= 0; i-- {
		c := schilds[i]
		if c.Name == entry.Name {
			continue
		}
		bts.WriteString("Vue.component('")
		bts.WriteString(c.Name)
		bts.WriteString("',")
		bts.WriteString(c.VueExportDefault)
		bts.WriteString(");\n")
	}
	bts.WriteString("Vue.component('")
	bts.WriteString(entry.Name)
	bts.WriteString("',")
	bts.WriteString(entry.VueExportDefault)
	bts.WriteString(");\n")

	//props:{resource:{type:Object, Require:true}}

	bts.WriteString("new Vue({el:'")
	bts.WriteString(fmt.Sprintf("div#%s.cc%s", elmounted, entry.Name))
	bts.WriteString("',")
	if entry.HasVuetify {
		bts.WriteString("vuetify: new Vuetify(),")
	}
	bts.WriteString("components:{")
	//bts.WriteString(entry.Name)
	bts.WriteString("},data:{},")
	bts.WriteString("})")

	btt.WriteString(fmt.Sprintf("<div id='%s' class='cc%s'>", elmounted, entry.Name))

	if entry.HasVuetify {
		btt.WriteString(fmt.Sprintf("<v-app><v-main><%s></%s></v-main></v-app>", entry.Name, entry.Name))
	} else {
		btt.WriteString(fmt.Sprintf("<%s></%s>", entry.Name, entry.Name))
	}
	btt.WriteString("</div>")

	R := &ResultOfVueBuilder{
		El:  btt.String(),
		Js:  bts.String(),
		Css: bty.String(),
	}
	return R, nil

	/*Vue.component('button-counter', {
	  template: '<button v-on:click="incrementHandler">{{ counter }}</button>',
	  data: function () {
	    return {
	      counter: 0
	    }
	  },
	  methods: {
	    incrementHandler: function () {
	      this.counter += 1
	      this.$emit('increment')
	    }
	  },
	})
	new Vue({
	  el: '#counter-event-example',
	  data: {
	    total: 0
	  },
	  methods: {
	    incrementTotal: function () {
	      this.total += 1
	    }
	  }
	})*/
}

func (v *VueBuilder) LoadAllComponents(dir string) error {
	files, err := ReadFilesInDir(dir, true, func(f os.FileInfo, d string) bool {
		if filepath.Ext(f.Name()) == ".vue" {
			return true
		}
		return false
	})

	if err != nil {
		return err
	}

	if v.Components == nil {
		v.Components = make([]*VueComponent, 0, len(files)*2)
	}

	for _, f := range files {
		vue := &VueComponent{}
		if err = vue.BuildFromComponentFile(f); err != nil {
			fmt.Println(fmt.Sprintf("[failed] load vue file [%s], %s", f, err.Error()))
			continue
		}

		for _, c := range v.Components {
			if c.Name == vue.Name {
				return fmt.Errorf("VueBuilder: file %s has duplicate component name %s with previous", f, vue.Name)
			}
		}
		v.Components = append(v.Components, vue)
	}

	return nil
}
