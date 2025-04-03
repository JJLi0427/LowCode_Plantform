package sformcompiler

import (
	"bytes"
	"fmt"
	"lowcode/core/control"
)

type SFormBuilder struct {
}

func (s *SFormBuilder) Generate(control *control.Control) error {
	if control.Input.View.Type != "sform" && control.Output.View.Type != "sform" {
		return nil
	}

	iformid, oformid := "ginputform", "goutputform"

	control.Tail.Scripts = append(control.Tail.Scripts,
		"/assets/js/sform/nearley-2.16.0.js",
		"/assets/js/sform/grammar.js",
		"/assets/js/sform/builder.js",
		"/assets/js/base/core_iobind.js")
	var btts bytes.Buffer
	if control.Input.View.Type == "sform" {
		front := ""
		if !control.Entrypoint.HasExeEntrypoint() && len(control.Input.View.Target) == 0 {
			front = "nobackserver"
		}
		btts.WriteString("function parseInputForm() { let parser = new nearley.Parser(grammar); let formText = `")
		btts.WriteString(control.Input.View.Inline_string)
		btts.WriteString("`; parser.feed(formText); let formJson = parser.results[0]; formHtml = builder(formJson);")
		btts.WriteString(fmt.Sprintf("document.querySelector('#%s').innerHTML = formHtml;} parseInputForm();sformOnSubmitConfig('#%s form', '%s');", iformid, iformid, front))
		control.SaveJS(btts.String(), false)

		btts.Reset()
		btts.WriteString(fmt.Sprintf("<div id='inputformcontainer'><div id='%s' class='fb-theme-default'></div></div>", iformid))
		control.Input.View.Inline_string = btts.String()
	}

	if control.Output.View.Type == "sform" {
		//btts.WriteString("function parseOutputForm() { let parser = new nearley.Parser(grammar); let formText = `")
		//btts.WriteString(control.Output.View.Inline_string)
		//btts.WriteString("`; parser.feed(formText); let formJson = parser.results[0]; formHtml = builder(formJson);")
		//btts.WriteString(fmt.Sprintf("document.querySelector('#%s').innerHTML = formHtml;} parseOutputForm();", oformid))
		btts.Reset()
		btts.WriteString(fmt.Sprintf("sformdispatchDataConfig('#outputformcontainer #%s');", oformid))
		btts.WriteString(fmt.Sprintf("dispatchData(`%s`);", control.Output.View.Inline_string))
		control.SaveJS(btts.String(), false)

		btts.Reset()
		btts.WriteString(fmt.Sprintf("<div id='outputformcontainer'><div id='%s' class='fb-theme-default'></div></div>", oformid))
		control.Output.View.Inline_string = btts.String()
	}

	return nil
}
