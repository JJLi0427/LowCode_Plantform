package cvuecompiler

import (
    "fmt"
    "onlinetools/core/control"
    "os"
    "path/filepath"
    "strings"
)

type componentSet struct {
    components map[string]*VueComponent
}

func newComponentSet() *componentSet {
    return &componentSet{
        components: make(map[string]*VueComponent),
    }
}

func (c *componentSet) Has(name string) bool {
    _, exists := c.components[name]
    return exists
}

func (c *componentSet) Append(component *VueComponent) bool {
    if !c.Has(component.Name) {
        c.components[component.Name] = component
        return true
    }
    return false
}

func (c *componentSet) Remove(name string) {
    delete(c.components, name)
}

func (c *componentSet) GetSlice() []*VueComponent {
    result := make([]*VueComponent, 0, len(c.components))
    for _, comp := range c.components {
        result = append(result, comp)
    }
    return result
}

func (c *componentSet) ToMap() map[string]*VueComponent {
    return c.components
}

func ReadFilesInDir(dir string, withPath bool, accept func(os.FileInfo, string) bool) ([]string, error) {
    info, err := os.Stat(dir)
    if err != nil || !info.IsDir() {
        return nil, fmt.Errorf("error [%s]: %v", dir, err)
    }

    var files []string
    err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() && accept(info, filepath.Dir(path)) {
            if withPath {
                files = append(files, path)
            } else {
                files = append(files, info.Name())
            }
        }
        return nil
    })

    return files, err
}

type ResultOfVueBuilder struct {
    El  string
    Js  string
    Css string
}

type VueBuilder struct {
    Components map[string]*VueComponent
}

func NewVueBuilder() *VueBuilder {
    return &VueBuilder{
        Components: make(map[string]*VueComponent),
    }
}

func (v *VueBuilder) hasVuetify(control *control.Control) bool {
    for _, script := range control.Head.Scripts {
        if strings.HasSuffix(strings.ToLower(script), "vuetify.js") {
            return true
        }
    }
    
    for _, script := range control.Tail.Scripts {
        if strings.HasSuffix(strings.ToLower(script), "vuetify.js") {
            return true
        }
    }
    
    return false
}

func (v *VueBuilder) Generate(control *control.Control) error {
    if control.Input.View.Type != "vue" && control.Output.View.Type != "vue" {
        return nil
    }

    hasVuetify := v.hasVuetify(control)

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

func (v *VueBuilder) FindAllChilds(component *VueComponent, childs *componentSet) error {
    for _, childName := range component.ChildComponents {
        childComp, exists := v.Components[childName]
        if !exists {
            return fmt.Errorf("找不到组件 %s", childName)
        }
        
        if childs.Append(childComp) {
            if err := v.FindAllChilds(childComp, childs); err != nil {
                return err
            }
        }
    }
    return nil
}

func createVueComponent(content string, filename string, isInput bool, hasVuetify bool) (*VueComponent, error) {
    var vueComp *VueComponent
    
    if isInput {
        vueComp = &VueComponent{HasVuetify: hasVuetify}
    } else {
        vueComp = &VueComponent{
            HasVuetify: hasVuetify,
            AddAddtionalDivs: []string{`<div id='selfinteraldataentryid' style='display:hidden;' @selfinteraldataentry='selfinteraldataentrymethod($event)'></div>`,},
            AddAddtionalMethods: []string{`selfinteraldataentrymethod(e){if (!e?.detail) {return}for (let key in e.detail) {this.$data[key] = e.detail[key]}}`},
            Scripts: `;vuedispatchDataConfig('#selfinteraldataentryid', 'selfinteraldataentry');`,
        }
    }

    if err := vueComp.BuildFromComponentFileContent([]byte(content)); err != nil {
        return nil, err
    }

    if len(vueComp.Name) == 0 && !vueComp.GenerateComponentNameByFileName(filename) {
        return nil, fmt.Errorf("faild to generate %s", filename)
    }

    if isInput {
        vueComp.Name = "i" + vueComp.Name
    } else {
        vueComp.Name = "o" + vueComp.Name
    }

    return vueComp, nil
}

func (v *VueBuilder) BuildIOVue(control *control.Control, elementId string, vuetify bool) (*ResultOfVueBuilder, error) {
    inputComponent, err := createVueComponent(control.Input.View.Inline_string, control.Input.View.Filename, true, vuetify)
    if err != nil {
        return nil, err
    }

    outputComponent, err := createVueComponent(control.Output.View.Inline_string, control.Output.View.Filename, false, vuetify)
    if err != nil {
        return nil, err
    }
	
    childs := newComponentSet()
    
    if len(v.Components) == 0 && (len(inputComponent.ChildComponents) > 0 || len(outputComponent.ChildComponents) > 0) {
        return nil, fmt.Errorf("VueBuilder: load all Vue components before building")
    }
    
    if err := v.FindAllChilds(inputComponent, childs); err != nil {
        return nil, err
    }
    
    if err := v.FindAllChilds(outputComponent, childs); err != nil {
        return nil, err
    }

    return v.buildComponentResult(inputComponent, outputComponent, childs, elementId, vuetify)
}

func (v *VueBuilder) BuildJS(vuecontent string, vuefilename string, elementId string, isInput bool, ignore map[string]*VueComponent, vuetify bool) (*ResultOfVueBuilder, error, map[string]*VueComponent) {
    vueComponent, err := createVueComponent(vuecontent, vuefilename, isInput, vuetify)
    if err != nil {
        return nil, err, nil
    }

    childs := newComponentSet()
    if len(vueComponent.ChildComponents) > 0 {
        if len(v.Components) == 0 {
            return nil, fmt.Errorf("VueBuilder: load all before build"), nil
        }
        
        if err := v.FindAllChilds(vueComponent, childs); err != nil {
            return nil, err, nil
        }
    }
	
    if ignore != nil {
        for name := range ignore {
            childs.Remove(name)
        }
    }

    if isInput {
        result, err := v.buildSingleComponent(vueComponent, childs, elementId)
        return result, err, childs.ToMap()
    } else {
        result, err := v.buildSingleComponent(vueComponent, childs, elementId)
        return result, err, childs.ToMap()
    }
}

func (v *VueBuilder) buildSingleComponent(component *VueComponent, childs *componentSet, elementId string) (*ResultOfVueBuilder, error) {
    var jsBuilder, cssBuilder, htmlBuilder strings.Builder

    jsBuilder.WriteString(component.Scripts)
    jsBuilder.WriteByte('\n')
    
    cssBuilder.WriteString(component.Styles)
    cssBuilder.WriteByte('\n')

    for _, child := range childs.GetSlice() {
        jsBuilder.WriteString(child.Scripts)
        jsBuilder.WriteByte('\n')
        
        cssBuilder.WriteString(child.Styles)
        cssBuilder.WriteByte('\n')
    }

    childComponents := childs.GetSlice()
    for i := len(childComponents) - 1; i >= 0; i-- {
        child := childComponents[i]
        if child.Name == component.Name {
            continue
        }
        jsBuilder.WriteString(fmt.Sprintf("Vue.component('%s', %s);\n", child.Name, child.VueExportDefault))
    }
	
    jsBuilder.WriteString(fmt.Sprintf("Vue.component('%s', %s);\n", component.Name, component.VueExportDefault))

    jsBuilder.WriteString(fmt.Sprintf("new Vue({el: 'div#%s.cc%s', ", elementId, component.Name))
    if component.HasVuetify {
        jsBuilder.WriteString("vuetify: new Vuetify(), ")
    }
    jsBuilder.WriteString("components:{}, data:{} })")

    htmlBuilder.WriteString(fmt.Sprintf("<div id='%s' class='cc%s'>", elementId, component.Name))
    if component.HasVuetify {
        htmlBuilder.WriteString(fmt.Sprintf("<v-app><v-main><%s></%s></v-main></v-app>", component.Name, component.Name))
    } else {
        htmlBuilder.WriteString(fmt.Sprintf("<%s></%s>", component.Name, component.Name))
    }
    htmlBuilder.WriteString("</div>")

    return &ResultOfVueBuilder{
        El:  htmlBuilder.String(),
        Js:  jsBuilder.String(),
        Css: cssBuilder.String(),
    }, nil
}

func (v *VueBuilder) buildComponentResult(inputComponent *VueComponent, outputComponent *VueComponent, childs *componentSet, elementId string, vuetify bool) (*ResultOfVueBuilder, error) {
    var jsBuilder, cssBuilder, htmlBuilder strings.Builder

    jsBuilder.WriteString(inputComponent.Scripts)
    jsBuilder.WriteByte('\n')
    jsBuilder.WriteString(outputComponent.Scripts)
    jsBuilder.WriteByte('\n')
    
    cssBuilder.WriteString(inputComponent.Styles)
    cssBuilder.WriteByte('\n')
    cssBuilder.WriteString(outputComponent.Styles)
    cssBuilder.WriteByte('\n')

    for _, child := range childs.GetSlice() {
        jsBuilder.WriteString(child.Scripts)
        jsBuilder.WriteByte('\n')
        
        cssBuilder.WriteString(child.Styles)
        cssBuilder.WriteByte('\n')
    }

    childComponents := childs.GetSlice()
    for i := len(childComponents) - 1; i >= 0; i-- {
        child := childComponents[i]
        if child.Name == inputComponent.Name || child.Name == outputComponent.Name {
            continue
        }
        jsBuilder.WriteString(fmt.Sprintf("Vue.component('%s', %s);\n", child.Name, child.VueExportDefault))
    }

    jsBuilder.WriteString(fmt.Sprintf("Vue.component('%s', %s);\n", inputComponent.Name, inputComponent.VueExportDefault))
    jsBuilder.WriteString(fmt.Sprintf("Vue.component('%s', %s);\n", outputComponent.Name, outputComponent.VueExportDefault))

    jsBuilder.WriteString(fmt.Sprintf("new Vue({el: 'div#%s.cc%s', ", elementId, inputComponent.Name))
    if vuetify {
        jsBuilder.WriteString("vuetify: new Vuetify(), ")
    }
    jsBuilder.WriteString("components:{}, data:{} })")

    htmlBuilder.WriteString(fmt.Sprintf("<div id='%s' class='cc%s'>", elementId, inputComponent.Name))
    if vuetify {
        htmlBuilder.WriteString(fmt.Sprintf("<v-app><v-main><%s></%s><%s></%s></v-main></v-app>", 
            inputComponent.Name, inputComponent.Name, outputComponent.Name, outputComponent.Name))
    } else {
        htmlBuilder.WriteString(fmt.Sprintf("<%s></%s><%s></%s>", 
            inputComponent.Name, inputComponent.Name, outputComponent.Name, outputComponent.Name))
    }
    htmlBuilder.WriteString("</div>")

    return &ResultOfVueBuilder{
        El:  htmlBuilder.String(),
        Js:  jsBuilder.String(),
        Css: cssBuilder.String(),
    }, nil
}

func (v *VueBuilder) LoadAllComponents(dir string) error {
    files, err := ReadFilesInDir(dir, true, func(f os.FileInfo, d string) bool {
        return filepath.Ext(f.Name()) == ".vue"
    })

    if err != nil {
        return err
    }

    if v.Components == nil {
        v.Components = make(map[string]*VueComponent, len(files))
    }

    for _, filePath := range files {
        vue := &VueComponent{}
        if err = vue.BuildFromComponentFile(filePath); err != nil {
            fmt.Printf("[faild] to load Vue [%s]: %s\n", filePath, err)
            continue
        }

        if _, exists := v.Components[vue.Name]; exists {
            return fmt.Errorf("VueBuilder: %s name %s reused", filePath, vue.Name)
        }
        
        v.Components[vue.Name] = vue
    }

    return nil
}