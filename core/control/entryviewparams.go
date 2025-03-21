package control

import (
	"bytes"
	"fmt"
	"strings"

	"code.cloudfoundry.org/bytefmt"
)

/*
paramsinview:
  # [optexe] paramoption in the excution, can be ignored if none. example: ./cmd.sh --optexe theinput
  # [:label] label in html form,
  # [type] form中输入元素的类型，支持的类别有：select,textarea,button,checkbox,file,image,password,radio,reset,text,color,date,datetime,datetime-local,email,month,number,range,search,tel,time,url,week
  # :{}, 表明该项只是在view中展示说明,例如标题，并不会被映射为entrypoint的参数
  - optexe:label{type[attr1='' attr2='']:v1,v2,v3,v4}
  - :{h1[attr1='' attr2='']:以下是工具的输入参数} #等价于 <h1 attr1=''  attr2=''>以下是工具的输入参数</h1>
  - --color:颜色{select[class="center" id='cc']:red,green,dark}
  - -info:information{checkbox:info1,info2,info3}
  - --pass:密码{passward}
  - :输入文本{textarea[pleaceholder="please input"]}.txt #.txt是可选的,表明entrypoint的输入参数是txt文件
  - :选择文件{file[accept="*.jpg"]}.jpg
*/

func trimQ(con string) string {
	con = strings.TrimRight(strings.TrimLeft(con, " "), " ")
	if l := len(con); l > 1 {
		if (con[0] == '"' && con[l-1] == '"') || (con[0] == '\'' && con[l-1] == '\'') {
			return con[1 : l-1]
		}
	}
	return con
}

type ViewParamBuilder struct {
	Ctrl *Control

	params              []*viewParam
	hasJsControlAttrs   bool
	onformchangejs      []byte
	// onsubmit_validatajs []byte
}

type viewParam struct {
	Optexe       string
	Label        string
	Elem         string
	ElemAttrs    string
	AttrControls map[string]string
	ElemContents []string
	FileExtType  string
	Node         string
	name         string
	id           string
	idx          int
}

func (v *viewParam) jsvalfnname() string {
	return fmt.Sprintf("formval%d", v.idx)
}
func (v *viewParam) jsvalfn() string {
	if len(v.Node) > 0 {
		return ""
	}

	var bts bytes.Buffer
	if v.Elem == "checkbox" || v.Elem == "radio" {
		bts.WriteString(fmt.Sprintf("function %s(){var marked=document.querySelectorAll('input[name=\"%s\"]:checked');var val='';for(var checked of marked){val+=checked.value+','};return val.replace(/[, ]*$/,'');}", v.jsvalfnname(), v.name))
	} else if v.Elem == "select" {
		bts.WriteString(fmt.Sprintf("function %s(){var selected=document.getElementById('%s');if(selected.selectedIndex>=0){return selected.options[selected.selectedIndex].value;}return '';}", v.jsvalfnname(), v.id))
	} else {
		bts.WriteString(fmt.Sprintf("function %s(){var selected=document.getElementById('%s');if(selected?.value){return selected.value;}return '';}", v.jsvalfnname(), v.id))
	}

	return bts.String()
}

func (v *viewParam) jsdependfn() string {
	var bts bytes.Buffer
	if dval, ok := v.AttrControls["_depend"]; ok && len(dval) > 0 {
		if v.Elem == "checkbox" || v.Elem == "radio" {
			bts.WriteString(fmt.Sprintf("var marked=document.querySelectorAll('input[name=\"%s\"]');if(%s){marked.forEach((elem)=>{elem.parentElement.parentElement.classList.remove('itemdisabled');elem.disabled=false})}else{marked.forEach((elem)=>{elem.parentElement.parentElement.classList.add('itemdisabled');elem.disabled=true})}", v.name, dval))
		} else {
			bts.WriteString(fmt.Sprintf("var selected%d=document.getElementById('%s');if(selected%d&&(%s)){selected%d.removeAttribute('disabled');selected%d.parentElement.classList.remove('itemdisabled')}else if(selected%d){selected%d.setAttribute('disabled','disabled');selected%d.parentElement.classList.add('itemdisabled')}", v.idx, v.id, v.idx, dval, v.idx, v.idx, v.idx, v.idx, v.idx))
		}
	}

	return bts.String()
}

func (v *viewParam) vparam(icnt int) string {
	v.idx = icnt
	if vv := strings.ReplaceAll(strings.TrimLeft(v.Optexe, "-"), " ", ""); len(vv) > 0 {
		v.name = vv
	} else {
		v.name = fmt.Sprintf("$%d", icnt)
	}

	if len(v.id) == 0 {
		if idx := strings.Index(v.ElemAttrs, "id="); idx == 0 && idx > 0 {
			var bts bytes.Buffer
			for _, c := range v.ElemAttrs[idx+3:] {
				if c == ' ' {
					break
				}
				bts.WriteRune(c)
			}
			v.id = trimQ(bts.String())
		} else {
			v.id = fmt.Sprintf("entry%s-%d", strings.TrimSpace(v.Elem), icnt)
			if v.Elem != "checkbox" && v.Elem != "radio" {
				v.ElemAttrs = fmt.Sprintf("id=\"%s\" %s", v.id, v.ElemAttrs)
			}
		}
	}

	return v.name
}
func (v *viewParam) getvparam() string {
	return v.name
}

func (v *viewParam) paramtype() string {
	if len(v.FileExtType) > 0 {
		if v.FileExtType == ".file" {
			return "file"
		}
		return fmt.Sprintf("file%s", v.FileExtType)
	}

	return "txt"
}

func checkAlphaChar(charVariable byte) bool {
	if (charVariable >= 'a' && charVariable <= 'z') || (charVariable >= 'A' && charVariable <= 'Z') {
		return true
	}
	return false
}

func (v *viewParam) parseElemAttrs(strAttrs string) string {
	//accept="*.jpg" _maxsize=4KB _depend="$1==red && $2~=info1"
	controls := make(map[string]string)
	var bts, cts bytes.Buffer
	pre, sq, dq := ' ', 0, 0
	for i, e := 0, len(strAttrs); i < e; {
		c := strAttrs[i]
		if pre == ' ' && sq%2 == 0 && dq%2 == 0 && c == '_' && i+1 < e && checkAlphaChar(strAttrs[i+1]) {
			//finded
			pr, chars, anchor, sq, dq := ' ', false, 0, 0, 0
			for j, m := range strAttrs[i:] {
				if m != ' ' {
					chars = true
				}
				if m == '=' && sq%2 == 0 && dq%2 == 0 && strAttrs[i+j-1] != '\\' && cts.Len() == 0 {
					anchor = j + 1
					cts.WriteString(strAttrs[i : i+j])
					chars = false
					pr = m
					continue
				} else if m == ' ' && chars && sq%2 == 0 && dq%2 == 0 && cts.Len() > 0 {
					controls[trimQ(cts.String())] = trimQ(strAttrs[i+anchor : i+j])
					cts.Reset()
					c = byte(m)
					i = i + j
					break
				} else if pr != '\\' && m == '"' {
					dq++
				} else if pr != '\\' && m == '\'' {
					sq++
				}
				pr = m
			}
			if cts.Len() > 0 && anchor > 0 {
				controls[trimQ(cts.String())] = trimQ(strAttrs[i+anchor:])
				cts.Reset()
				break
			}
		} else if pre != '\\' && c == '"' {
			dq++
		} else if pre != '\\' && c == '\'' {
			sq++
		}
		bts.WriteByte(c)

		pre = rune(strAttrs[i])
		i++
	}

	if len(controls) > 0 {
		v.AttrControls = controls
	}

	return bts.String()
}

func (v *viewParam) parse(str string) error {
	str = strings.TrimRight(strings.TrimLeft(str, " \t"), " \t")
	if len(str) > 2 && str[0] == '<' && str[len(str)-1] == '>' {
		v.Node = str
		return nil
	}

	before, brace, after, s, e, dq, sq, pre := "", "", "", -1, -1, 0, 0, ' '
	for i, c := range str {
		if dq%2 == 0 && sq%2 == 0 && c == '{' && pre != '\\' {
			s = i
		} else if dq%2 == 0 && sq%2 == 0 && c == '}' && pre != '\\' {
			e = i
		} else if c == '"' && pre != '\\' {
			dq++
		} else if c == '\'' && pre != '\\' {
			sq++
		}
		pre = c
	}

	if s < e && s >= 0 {
		before = strings.TrimSpace(str[:s])
		brace = str[s+1 : e]
		after = strings.TrimSpace(str[e+1:])
	} else if s < 0 {
		before = str
	}

	if idx := strings.Index(before, ":"); idx >= 0 {
		v.Optexe = before[:idx]
		v.Label = trimQ(before[idx+1:])
	}

	if len(after) > 1 && after[0] == '.' {
		v.FileExtType = after
	}

	//{h1[attr1='' attr2='']:以下是工具的输入参数}
	if len(brace) > 0 {
		var bts bytes.Buffer
		elem, attrs, content := "", "", ""
		pre, sq, dq, l := byte(' '), 0, 0, false
		for i, c := range []byte(brace) {
			if sq%2 == 0 && dq%2 == 0 && c == '[' {
				l = true
				elem = bts.String()
				bts.Reset()
				pre = c
				continue
			} else if sq%2 == 0 && dq%2 == 0 && c == ']' && l {
				attrs = bts.String()
				bts.Reset()
				pre = c
				continue
			} else if sq%2 == 0 && dq%2 == 0 && c == ':' {
				if len(elem) == 0 {
					elem = bts.String()
					bts.Reset()
				}
				if l && len(attrs) == 0 {
					panic(fmt.Sprintf("failed to parse entrypoint paramsinview[%s]", str))
				}
				content = brace[i+1:]
				break
			} else if pre != '\\' && c == '"' {
				dq++
			} else if pre != '\\' && c == '\'' {
				sq++
			} else if pre == '\\' && (c == '\'' || c == '"') {
				bts.Truncate(bts.Len() - 1)
				bts.WriteByte(c)
			}
			bts.WriteByte(c)
			pre = c
		}
		if len(elem) == 0 && len(attrs) == 0 && len(content) == 0 && bts.Len() > 0 {
			elem = bts.String()
			bts.Reset()
		}

		v.Elem = elem
		v.ElemAttrs = v.parseElemAttrs(attrs)

		if len(content) > 0 {
			var bts bytes.Buffer
			ccc := []byte(content)
			pre, sq, dq := byte(' '), 0, 0
			for _, c := range ccc {
				if sq%2 == 0 && dq%2 == 0 && c == ',' {
					v.ElemContents = append(v.ElemContents, bts.String())
					bts.Reset()
					pre = c
					continue
				} else if pre != '\\' && c == '"' {
					dq++
				} else if pre != '\\' && c == '\'' {
					sq++
				} else if pre == '\\' && (c == '\'' || c == '"') {
					bts.Truncate(bts.Len() - 1)
					bts.WriteByte(c)
				}
				bts.WriteByte(c)
				pre = c
			}

			if bts.Len() > 0 && len(strings.TrimSpace(bts.String())) > 0 {
				v.ElemContents = append(v.ElemContents, bts.String())
			}
		}
	}

	if len(v.Optexe) == 0 && len(v.Label) == 0 && len(v.Elem) > 0 {
		s := strings.Join(v.ElemContents, ",")
		v.Node = fmt.Sprintf("<%s %s>%s</%s>", v.Elem, v.ElemAttrs, s, v.Elem)
	}

	return nil
}

func (v *ViewParamBuilder) parse(vstr []string) error {
	v.params = nil
	for _, s := range vstr {
		p := &viewParam{}
		if err := p.parse(s); err != nil {
			return err
		}
		v.params = append(v.params, p)

		if !v.hasJsControlAttrs {
			//internal [_depend] attr
			if _, ok := p.AttrControls["_depend"]; ok {
				v.hasJsControlAttrs = true
			}
		}
	}

	return nil
}

func (v *ViewParamBuilder) buildStdin() ([]ParamMaps, error) {
	var pmaps []ParamMaps
	for i, param := range v.params {
		if len(param.Node) > 0 || (len(param.Optexe) == 0 && len(param.Label) == 0) {
			continue
		}

		required, def := true, ""
		if _, ok := param.AttrControls["_depend"]; ok {
			required = false
		}
		if df, ok := param.AttrControls["_default"]; ok {
			def = df
			if len(df) == 0 {
				required = false
			}
		}

		pmaps = append(pmaps, ParamMaps{Viewparam: param.vparam(i),
			Exeopt:    param.Optexe,
			Paramtype: param.paramtype(),
			Required:  required,
			Default:   def,
		})
	}

	return pmaps, nil
}

// for cmd: _validate
func (v *ViewParamBuilder) buildJSFormDataValidate() (string, error) {
	var bts bytes.Buffer
	bts.WriteString("function ValidateFormData(form_data){")
	for _, param := range v.params {
		if len(param.Node) > 0 {
			continue
		}
		bts.WriteString("\n")
		bts.WriteString(param.jsvalfn())
		bts.WriteString(fmt.Sprintf("var $%d=%s();\n", param.idx, param.jsvalfnname()))
		//bts.WriteString(fmt.Sprintf("\nconsole.log('%s',%s());", param.name, param.jsvalfnname()))
	}

	cnt := 0
	for i, param := range v.params {
		if fn, ok := param.AttrControls["_validate"]; ok && len(fn) > 0 {
			idsubfix := ""
			if len(param.ElemContents) > 1 {
				idsubfix = "-fk"
			}

			if param.Elem == "text" {
				cnt += 1
				bts.WriteString(fmt.Sprintf("var text%d=document.getElementById('%s%s'); var _valifn%d = %s; var ret = true; if(!form_data.has('%s')){ret=true;}else if (typeof _valifn%d == 'function'){ ret = _valifn%d(form_data.get('%s')); } else if (isRegularExpression(_valifn%d)){ ret=_valifn%d.test(form_data.get('%s'));if(!ret){ret='invalid';}} if (typeof ret == 'string'){ if (ret.length > 0) { text%d.dataset.validate=ret; ret = false; }else{ ret = true; } } if (!ret && !text%d.classList.contains('invalid')) { text%d.classList.add('invalid'); text%d.parentElement.classList.add('invalid-tips'); }else{ text%d.classList.remove('invalid');text%d.parentElement.classList.remove('invalid-tips');}",
					i, param.id, idsubfix, i, fn, param.vparam(i), i, i, param.vparam(i), i, i, param.vparam(i), i, i, i, i, i, i))
				bts.WriteString("if (ret == false){return ret;}")
			}
		}
	}

	bts.WriteString("\n return true; }")

	if cnt == 0 {
		return "", nil
	}

	return bts.String(), nil
}

// for cmd: _value
func (v *ViewParamBuilder) buildJSFormDataRewriteValue() (string, error) {
	var bts bytes.Buffer
	bts.WriteString("function ReWriteFormData(form_data){")
	for _, param := range v.params {
		if len(param.Node) > 0 {
			continue
		}
		bts.WriteString("\n")
		bts.WriteString(param.jsvalfn())
		bts.WriteString(fmt.Sprintf("var $%d=%s();\n", param.idx, param.jsvalfnname()))
		//bts.WriteString(fmt.Sprintf("\nconsole.log('%s',%s());", param.name, param.jsvalfnname()))
	}

	cnt := 0
	for i, param := range v.params {
		if fn, ok := param.AttrControls["_value"]; ok && len(fn) > 0 {
			cnt += 1
			vari := fmt.Sprintf("fn_%s_%s", "value", param.vparam(i))
			val := fmt.Sprintf("if(form_data.has('%s')){let %s=%s;form_data.set('%s', %s(form_data.get('%s')))}\n",
				param.vparam(i), vari, fn, param.vparam(i), vari, param.vparam(i))
			bts.WriteString(val)
		}
	}

	bts.WriteString("\n return form_data; }")

	if cnt == 0 {
		return "", nil
	}

	return bts.String(), nil
}

func (v *ViewParamBuilder) buildFormViewOnChangeJSEvent() (string, error) {
	var bts bytes.Buffer
	//bts.WriteString(fmt.Sprintf("function onformchange('%s'){", formid))
	bts.WriteString("function onformchange(formid){")
	if v.hasJsControlAttrs {
		for _, param := range v.params {
			if len(param.Node) > 0 {
				continue
			}
			bts.WriteString("\n")
			bts.WriteString(param.jsvalfn())
			bts.WriteString(fmt.Sprintf("var $%d=%s();", param.idx, param.jsvalfnname()))
			//bts.WriteString(fmt.Sprintf("\nconsole.log('%s',%s());", param.name, param.jsvalfnname()))
		}
	}

	for _, param := range v.params {
		if len(param.AttrControls) == 0 {
			continue
		}
		//for file maxsize
		if val, ok := param.AttrControls["_maxsize"]; ok && param.Elem == "file" && len(val) > 0 {
			if nbytes, err := bytefmt.ToBytes(val); err == nil {
				bts.WriteString(fmt.Sprintf("var ifile%d=document.getElementById('%s');", param.idx, param.id))
				bts.WriteString(fmt.Sprintf("if(ifile%d.files.length){if(ifile%d.files[0].size>%d){alert(`文件最大不能超过%s!\nthe file's max-size must be less than %s!`);ifile%d.value='';}}", param.idx, param.idx, nbytes, val, val, param.idx))
			} else {
				panic(err)
			}
		}

		if v.hasJsControlAttrs {
			//for _depend
			bts.WriteString(param.jsdependfn())
		}
	}
	bts.WriteString("}")

	if len(v.onformchangejs) > 0 {
		bts.WriteString(fmt.Sprintf("document.ready(()=>{%s});", v.onformchangejs))
		v.onformchangejs = nil
	}

	bts.WriteString("document.ready(onformchange);")

	if str, err := v.buildJSFormDataRewriteValue(); len(str) > 0 && err == nil {
		bts.WriteString(str)
	}
	if str, err := v.buildJSFormDataValidate(); len(str) > 0 && err == nil {
		bts.WriteString(str)
	}

	return bts.String(), nil
}

func (v *ViewParamBuilder) buildFormView(target, method string) (*ViewControl, error) {
	var bts bytes.Buffer
	var scripts bytes.Buffer
	inline := false
	v.Ctrl.Head.Scripts = append(v.Ctrl.Head.Scripts, "/assets/js/base/entrypointform.js")

	bts.WriteString(fmt.Sprintf("\n<form id=\"entrypointform\" onchange=\"onformchange('entrypointform')\" action=\"%s\" method=\"%s\">", target, method))
	for i, param := range v.params {
		//[type] form中输入元素的类型，支持的类别有：select,textarea,button,checkbox,file,image,password,radio,reset,text,color,date,datetime,datetime-local,email,month,number,range,search,tel,time,url,week
		if _, ok := param.AttrControls["_inline"]; ok && !inline {
			inline = true
			bts.WriteString(fmt.Sprintf("\n<div class=\"formitem formrow formrow%d\">", i))
		} else if _, ok = param.AttrControls["_inline"]; (!ok || i == len(v.params)-1) && inline {
			inline = false
			bts.WriteString("\n</div>")
		}

		if len(param.Node) > 0 {
			bts.WriteString("\n" + param.Node)
			continue
		}
		deval := ""
		if val, ok := param.AttrControls["_default"]; ok && len(val) > 0 {
			deval = val
		}

		tooltipcla, tooltip := "", ""
		if val, ok := param.AttrControls["_tooltips"]; ok && len(val) > 0 {
			tooltipcla = "tooltip-container"
			tooltip = "data-tooltip='" + val + "'"
		}

		if inline {
			bts.WriteString(fmt.Sprintf("\n<div class=\"%s forminlineitem formitem%d\" %s>", tooltipcla, i, tooltip))
		} else {
			bts.WriteString(fmt.Sprintf("\n<div class=\"%s formitem formitem%d\" %s>", tooltipcla, i, tooltip))
		}

		if param.Elem == "select" {
			if len(param.Label) > 0 {
				bts.WriteString(fmt.Sprintf("\n<label class=\"formlabel formlabelselect \">%s</label>", param.Label))
			}
			bts.WriteString(fmt.Sprintf("\n<select %s name=\"%s\">", param.ElemAttrs, param.getvparam()))
			selected := false
			for _, con := range param.ElemContents {
				checked := ""
				if !selected && len(deval) > 0 && strings.Contains(con, deval) {
					checked = "selected"
					selected = true
				}
				if idx := strings.Index(con, "/"); idx > 0 {
					bts.WriteString(fmt.Sprintf("\n<option %s value=\"%s\">%s</option>", checked, trimQ(con[:idx]), trimQ(con[idx+1:])))
				} else {
					bts.WriteString(fmt.Sprintf("\n<option %s value=\"%s\">%s</option>", checked, trimQ(con), trimQ(con)))
				}
			}
			bts.WriteString("\n</select>")
		} else if param.Elem == "textarea" {
			if len(param.Label) > 0 {
				bts.WriteString(fmt.Sprintf("\n<label class=\"formlabel formlabeltextarea\">%s</label>", param.Label))
			}
			bts.WriteString(fmt.Sprintf("\n<textarea %s name=\"%s\">%s</textarea>", param.ElemAttrs, param.getvparam(), deval))
		} else if param.Elem == "input" {
			if len(deval) > 0 {
				deval = "value=\"" + deval + "\""
			}
			bts.WriteString(fmt.Sprintf("\n<label class=\"formlabel formlabelinput\">%s</label><input %s name=\"%s\" %s></input>", param.Label, param.ElemAttrs, param.getvparam(), deval))
		} else if param.Elem == "checkbox" {
			if len(param.Label) > 0 {
				bts.WriteString(fmt.Sprintf("\n<label class=\"formlabel formlabelinput formlabelcheckbox\">%s</label>", param.Label))
			}
			for _, con := range param.ElemContents {
				checked := ""
				if len(deval) > 0 && strings.Contains(con, deval) {
					checked = "checked"
				}
				if idx := strings.Index(con, "/"); idx > 0 {
					bts.WriteString(fmt.Sprintf("\n<label class='checkboxcontainer'><input type=\"checkbox\" %s name=\"%s\" value=\"%s\" %s /> %s </label>", param.ElemAttrs, param.getvparam(), trimQ(con[:idx]), checked, trimQ(con[idx+1:])))
				} else {
					bts.WriteString(fmt.Sprintf("\n<label class='checkboxcontainer'><input type=\"checkbox\" %s name=\"%s\" value=\"%s\" %s /> %s </label>", param.ElemAttrs, param.getvparam(), trimQ(con), checked, trimQ(con)))
				}
			}
		} else if param.Elem == "radio" {
			if len(param.Label) > 0 {
				bts.WriteString(fmt.Sprintf("\n<label class=\"formlabel formlabelinput formlabelradio\">%s</label>", param.Label))
			}
			for _, con := range param.ElemContents {
				checked := ""
				if len(deval) > 0 && strings.Contains(con, deval) {
					checked = "checked"
				}
				if idx := strings.Index(con, "/"); idx > 0 {
					bts.WriteString(fmt.Sprintf("\n<label class='radiocontainer'><input type=\"radio\" %s name=\"%s\" value=\"%s\" %s /> %s </label>", param.ElemAttrs, param.getvparam(), trimQ(con[:idx]), checked, trimQ(con[idx+1:])))
				} else {
					bts.WriteString(fmt.Sprintf("\n<label class='radiocontainer'><input type=\"radio\" %s name=\"%s\" value=\"%s\" %s /> %s </label>", param.ElemAttrs, param.getvparam(), trimQ(con), checked, trimQ(con)))
				}
			}
		} else if param.Elem == "range" {
			if scripts.Len() == 0 {
				scripts.WriteString("function onirangeinput(id){document.getElementById('label'+id).innerHTML = document.getElementById(id).value;}")
			}
			oninput := fmt.Sprintf("onirangeinput('%s')", param.id)
			scripts.WriteString(oninput)
			scripts.WriteByte(';')
			if len(deval) > 0 {
				deval = "value=\"" + deval + "\""
			}
			bts.WriteString(fmt.Sprintf("\n<label class=\"formlabel formlabelinput formlabel%s\">%s</label>\n<input type=\"%s\" %s name=\"%s\" %s oninput=\"%s\"/><label class=\"labelrangevalue\" id=\"label%s\"></label>", param.Elem, param.Label, param.Elem, param.ElemAttrs, param.getvparam(), deval, oninput, param.id))
		} else if param.Elem == "text" && len(param.ElemContents) > 1 {
			if len(param.Label) > 0 {
				bts.WriteString(fmt.Sprintf("\n<label class=\"formlabel formlabelinput formlabel%s\">%s</label>", param.Elem, param.Label))
			}

			depval, dval := "", ""
			if len(deval) > 0 {
				for _, con := range param.ElemContents {
					disp, val := con, con
					if idx := strings.Index(con, "/"); idx > 0 {
						disp = con[idx+1:]
						val = con[:idx]
					}
					if disp == deval || val == deval {
						dval = "value='" + val + "'"
						depval = "value='" + disp + "'"
						break
					}
				}
			}

			bts.WriteString("\n<div class='selectedinputtext'>")
			bts.WriteString(fmt.Sprintf("\n<input type='text' id='%s' class='displaynone' name='%s' %s>", param.id, param.getvparam(), dval))
			bts.WriteString(fmt.Sprintf("\n<input type='text' id='%s-fk' %s onblur=\"setTimeout(()=>{selectedinputtext_hidelist('dataid%s')},200)\" onclick=\"selectedinputtext_expandlist('dataid%s')\"  oninput=\"selectedinputtext_onchange(event, 'dataid%s', '%s')\" autocomplete=\"off\" %s>", param.id, param.ElemAttrs, param.getvparam(), param.getvparam(), param.getvparam(), param.id, depval))
			bts.WriteString(fmt.Sprintf("\n<div class='selectedinputtextlist displaynone' id='dataid%s' > <ul>", param.getvparam()))
			for _, con := range param.ElemContents {
				disp, val := con, con
				if idx := strings.Index(con, "/"); idx > 0 {
					disp = con[idx+1:]
					val = con[:idx]
				}
				bts.WriteString(fmt.Sprintf("\n<li onclick=\"selectedinputtext_selectli(event, '%s', '%s');selectedinputtext_hidelist('dataid%s')\">%s</li>", val, param.id, param.getvparam(), disp))
			}
			bts.WriteString("</ul></div></div>")

			var jsb bytes.Buffer
			jsb.WriteString(fmt.Sprintf("\nselectedinputtext_mutationfn('%s', '%s-fk');", param.id, param.id))
			v.onformchangejs = append(v.onformchangejs, jsb.Bytes()...)
		} else {
			if len(param.ElemContents) > 0 {
				bts.WriteString(fmt.Sprintf("\n<label class=\"formlabel formlabelinput formlabel%s\">%s</label>", param.Elem, param.Label))
				for _, con := range param.ElemContents {
					if idx := strings.Index(con, "/"); idx > 0 {
						bts.WriteString(fmt.Sprintf("\n<label class='%scontainer'><input type=\"%s\" %s name=\"%s\" value=\"%s\" /> %s </label>", param.Elem, param.Elem, param.ElemAttrs, param.getvparam(), trimQ(con[:idx]), trimQ(con[idx+1:])))
					} else {
						bts.WriteString(fmt.Sprintf("\n<input type=\"%s\" %s name=\"%s\" value=\"%s\" />", param.Elem, param.ElemAttrs, param.getvparam(), trimQ(con)))
					}
				}
			} else {
				if len(deval) > 0 {
					deval = "value=\"" + deval + "\""
				}
				bts.WriteString(fmt.Sprintf("\n<label class=\"formlabel formlabelinput formlabel%s\">%s</label>\n<input type=\"%s\" %s name=\"%s\" %s />", param.Elem, param.Label, param.Elem, param.ElemAttrs, param.getvparam(), deval))
			}
		}
		bts.WriteString("\n</div>")
	}

	hassubmit := false
	for _, param := range v.params {
		if param.Elem == "submit" || param.Elem == "button" {
			hassubmit = true
			break
		}
	}
	if !hassubmit {
		bts.WriteString("\n<div class=\"formitem\">")
		bts.WriteString("\n<label class=\"formlabel formlabelinput formlabelsubmit\"><input type=\"submit\" value=\"提交\"/></label>")
		bts.WriteString("\n</div>")
	}

	bts.WriteString("\n</form>")

	if changeevent, _ := v.buildFormViewOnChangeJSEvent(); scripts.Len() > 0 || len(changeevent) > 0 {
		bts.WriteString(fmt.Sprintf("\n<script>%s\n%s</script>", scripts.String(), changeevent))
	}

	view := &ViewControl{Type: "html", Target: target}
	view.Inline_string = bts.String()

	//fmt.Println(view.Inline_string)

	return view, nil
}

func (v *ViewParamBuilder) buildStdout() ([]ParamMaps, error) {
	for i, param := range v.params {
		if len(param.Optexe) == 0 {
			v.params[i].Optexe = "__wholecontent__"
		}
	}

	params, err := v.buildStdin()
	if err == nil {
		for i, param := range params {
			if len(param.Exeparam) == 0 {
				params[i].Exeparam = param.Exeopt
				params[i].Exeopt = ""
			}
		}
	}

	return params, err
}

func (v *ViewParamBuilder) BuildIOControl(vstr []string, target, method string) (*IOControl, error) {
	if err := v.parse(vstr); err != nil {
		return nil, err
	}

	input := &IOControl{Type: "form"}
	if stdin, err := v.buildStdin(); err != nil {
		return nil, err
	} else {
		input.Stdin = stdin
	}

	if view, err := v.buildFormView(target, method); err != nil {
		return nil, err
	} else {
		input.View = *view
	}

	return input, nil
}

func (v *ViewParamBuilder) buildOutputView() (*ViewControl, string, error) {
	var bts bytes.Buffer
	var btscript bytes.Buffer

	if len(v.params) > 0 {
		bts.WriteString(`<div id="entryoutput" class="entrypointview">`)
		btscript.WriteString(`window.dispatchData = function(payload){var entryoutview = document.getElementById('entryoutput');entryoutview.innerHTML = '';`)
	}

	for i, param := range v.params {
		if len(param.Node) > 0 {
			btscript.WriteString(fmt.Sprintf("entryoutview.appendChild(parseHtmlElement('%s'));", param.Node))
			continue
		}
		//支持的标签有：div, img, video, audio, h1-h6, a等
		node := fmt.Sprintf("<%s class='entryoutview' %s>%s</%s>", param.Elem, param.ElemAttrs, strings.Join(param.ElemContents, ","), param.Elem)
		if param.Elem == "img" {
			node = fmt.Sprintf("<%s class='entryoutview' %s data-content='%s'>", param.Elem, param.ElemAttrs, strings.Join(param.ElemContents, ","))
		}
		btscript.WriteString(fmt.Sprintf("if(payload?.%s || '%s' == '__wholecontent__'){var elem%d = parseHtmlElement(`%s`);", param.getvparam(), param.getvparam(), i, node))
		loc, content := "innerHTML", strings.Join(param.ElemContents, ",")
		if param.Elem == "img" || param.Elem == "video" || param.Elem == "audio" {
			loc = "src"
		} else if param.Elem == "a" {
			loc = "href"
		}
		btscript.WriteString(fmt.Sprintf("if(payload?.%s){elem%d.%s=payload.%s;}else if ('%s' == '__wholecontent__'){elem%d.%s=payload;}", param.getvparam(), i, loc, param.getvparam(), param.getvparam(), i, loc))
		btscript.WriteString(fmt.Sprintf("if('%s'.length > 0 && '%s' != 'innerHTML'){elem%d.innerHTML='%s';}else if(payload?.%s && '%s' == 'href'){elem%d.innerHTML=parseUrlName(payload.%s);}", content, loc, i, content, param.getvparam(), loc, i, param.getvparam()))
		btscript.WriteString(fmt.Sprintf("var cdiv%d = document.createElement('div'); cdiv%d.className='entryoutitem';", i, i))
		if len(param.Label) > 0 {
			btscript.WriteString(fmt.Sprintf("var clabel%d = document.createElement('label');clabel%d.innerHTML=`%s`;clabel%d.className='entryoutlabel'; cdiv%d.appendChild(clabel%d);", i, i, param.Label, i, i, i))
		}
		btscript.WriteString(fmt.Sprintf("cdiv%d.appendChild(elem%d);entryoutview.appendChild(cdiv%d);}\n", i, i, i))
	}

	btscript.WriteString("}")
	bts.WriteString("</div>")

	view := &ViewControl{Type: "html"}
	view.Inline_string = bts.String()

	return view, btscript.String(), nil
}

func (v *ViewParamBuilder) BuildOuputIOControl(vstr []string) (*IOControl, string, error) {
	if err := v.parse(vstr); err != nil {
		return nil, "", err
	}

	output := &IOControl{Type: "kvstr,json"}
	if stdout, err := v.buildStdout(); err != nil {
		return nil, "", err
	} else {
		output.Stdout = stdout
	}

	if view, script, err := v.buildOutputView(); err != nil {
		return nil, "", err
	} else {
		output.View = *view
		return output, script, nil
	}
}
