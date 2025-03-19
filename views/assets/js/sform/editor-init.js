var langTools = ace.require("ace/ext/language_tools");
var snippetManager = ace.require("ace/snippets").snippetManager;

// var htmlResultEditor = ace.edit("html-result-editor");
// htmlResultEditor.setTheme("ace/theme/monokai");
// htmlResultEditor.session.setMode("ace/mode/html");

var editor = ace.edit("editor");
editor.setOptions({
    enableBasicAutocompletion:false,
    enableSnippets: true,
    enableLiveAutocompletion: true
});
editor.setTheme("ace/theme/monokai");
editor.session.setMode("ace/mode/formbuilder");
editor.session.on('change', debounce(function() {
    parseForm();
}, 600));

var keywords = "form;name;header;method;autocomplete;action;section;field;type;required;placeholder;icon;value;hint;label;validations;maxlength;minlength;max;min;end validations;options;end options;end field".split(";");
var vars = null;
var consts = "post;get;on;off;textarea;textbox;phone;email;url;number;date;time;hidden;search;password;checkboxgroup;radiogroup;submit;select;combo;datepicker;checked;selected;address;address-google".split(";");

// create a completer object with a required callback function:
var formBuilderCompleter = {
    getCompletions: function(editor, session, pos, prefix, callback) {
        var completions = [];
        keywords.forEach(function (kw) {
            completions.push({value:kw, scrore:1001, meta:"Keyword"})
        });
        consts.forEach(function (c) {
            completions.push({value:c, score:1000, meta:"Enum"})
        });
        completions.push({value:"f-checkboxes", score:999, meta:"Snippet"});
        completions.push({value:"f-email", score:999, meta:"Snippet"});
        completions.push({value:"f-gaddress", score:999, meta:"Snippet"});
        completions.push({value:"f-name", score:999, meta:"Snippet"});
        completions.push({value:"f-password", score:999, meta:"Snippet"});
        completions.push({value:"f-phone", score:999, meta:"Snippet"});
        completions.push({value:"f-radio", score:999, meta:"Snippet"});
        completions.push({value:"f-select", score:999, meta:"Snippet"});
        completions.push({value:"f-submit", score:999, meta:"Snippet"});
        completions.push({value:"f-text", score:999, meta:"Snippet"});
        completions.push({value:"f-textarea", score:999, meta:"Snippet"});
        completions.push({value:"f-url", score:999, meta:"Snippet"});
        callback(null, completions);
    }
};
langTools.setCompleters([]);
// finally, bind to langTools:
langTools.addCompleter(formBuilderCompleter);



var snippets = [];// snippetManager.parseSnippetFile("snippet test\n  TEST!");
snippets.push({
    content: `field\nname \${1:fieldname}\ntype textbox\nrequired\nplaceholder \${2:placeholder_text}\nvalue \${3:default_value}\nlabel \${4:field_label}\nhint \${5:field_hint}\nvalidations\nminlength 1\nmaxlength 80\nend validations\nend field`,
    name: "Text Field",
    tabTrigger: "f-text"
});
snippets.push({
    content: `field\nname \${1:email}\ntype email\nicon fas fa-envelope\nrequired\nplaceholder \${2:Please enter your email}\nvalue \${3:default_value}\nlabel \${4:Email}\nhint \${5:We hate spam too. We will never spam you.}\nvalidations\nminlength 1\nmaxlength 200\nend validations\nend field`,
    name: "Email Field",
    tabTrigger: "f-email"
});
snippets.push({
    content: `field\nname \${1:phone}\ntype phone\nicon fas fa-phone\nrequired\nplaceholder \${2:Please enter your phone}\nvalue \${3:default_value}\nlabel \${4:Phone}\nhint \${5:field_hint}\nend field`,
    name: "Phone Field",
    tabTrigger: "f-phone"
});
snippets.push({
    content: `field\nname \${1:website}\ntype url\nicon fas fa-globe\nrequired\nplaceholder \${2:placeholder_text}\nvalue \${3:default_value}\nlabel \${4:Website}\nhint \${5:field_hint}\nvalidations\nminlength 1\nmaxlength 300\nend validations\nend field`,
    name: "URL Field",
    tabTrigger: "f-url"
});
snippets.push({
    content: `field\nname \${1:password}\ntype password\nicon fas fa-lock\nrequired\nplaceholder \${2:Choose your password}\nlabel \${4:Password}\nhint \${5:Choose a strong password that you can remember}\nvalidations\nminlength 6\nmaxlength 300\nend validations\nend field`,
    name: "Password Field",
    tabTrigger: "f-password"
});
snippets.push({
    content: `field\nname \${1:message}\ntype textarea\nicon fas fa-comment-alt\nrequired\nplaceholder \${2:We would love to hear what you have to say}\nvalue \${3:default_value}\nlabel \${4:Message}\nhint \${5:field_hint}\nvalidations\nminlength 1\nmaxlength 1500\nend validations\nend field`,
    name: "Textarea Field",
    tabTrigger: "f-textarea"
});
snippets.push({
    content: `field\nname name\nicon fas fa-user-tie\ntype textbox\nrequired\nplaceholder What is your name?\nlabel Name\nvalidations\nminlength 1\nmaxlength 200\nend validations\nend field`,
    name: "Name Field",
    tabTrigger: "f-name"
});
snippets.push({
    content: `field\ntype submit\nlabel Submit\nend field`,
    name: "Submit Button",
    tabTrigger: "f-submit"
});
snippets.push({
    content: `field\nname \${1:color}\ntype select\nicon \${2:fas fa-palette}\noptions\n\${3:"Red" | "#ff0000"\n"Green" | "#00ff00" | selected\n"Blue" | "#0000ff"\n"Orange" | "#ffa500"\n"Purple" | "#800080"\n"White" | "#ffffff"}\nend options\n\${4:required}\nlabel \${5:What is your favorite color?}\nend field`,
    name: "Select Option",
    tabTrigger: "f-select"
});
snippets.push({
    content: `field\nname \${1:age}\ntype radiogroup\nicon \${2:fas fa-birthday-cake}\noptions\n\${3:"Less than 18" | "<18" | checked\n"18-25" | "18-25"\n"26-35" | "26-35"\n"36-45" | "36-45"\n"46-65" | "46-65"\n"66+" | ">66"}\nend options\n\${4:required}\nlabel \${5:How old are you?}\nend field`,
    name: "Radio Group",
    tabTrigger: "f-radio"
});
snippets.push({
    content: `field\nname \${1:platforms}\ntype checkboxgroup\nicon \${2:fas fa-desktop}\noptions\n\${3:"Windows"\n"Linux"\n"Mac"\n"Other"}\nend options\n\${4:required}\nlabel \${5:Which platforms do you use?}\nend field`,
    name: "Checkbox Group",
    tabTrigger: "f-checkboxes"
});
snippets.push({
    content: `field\nname address\ntype address-google\nlabel Address\nicon fas fas fa-map-marker-alt\nrequired\nplaceholder 4 Penny Lane, Liverpool, England\nvalidations\nmaxlength 250\nminlength 2\nend validations\nend field`,
    name: "Google Address Field",
    tabTrigger: "f-gaddress"
});


snippetManager.register(snippets, "formbuilder");


function initAddButtons() {
   document.querySelectorAll(".add-item").forEach(el => {
       el.onclick = e => {
           addSnippet(e.target.dataset.fieldname);
       }
   }) ;
}

initAddButtons();

let sharedHtml;
function getHtml() {
    sharedHtml = `<!-- this should be placed in the HEAD -->
<link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/danbars/form-builder@0.1/dist/themes/default.imports.css">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/danbars/form-builder@0.1/dist/themes/default.css">
<div id="form" class="fb-theme-default">
${formHtml}
</div>`;
    document.querySelector("#html-result-editor code").innerHTML = escapeHtml(sharedHtml);
    // document.querySelector("#copyToClipboard").dataset.dataClipboardText = formHtml;
    uglipop({class:'my-styling-class', //styling class for Modal
        source:'div', //'div' instead of 'html'
        content:'html-result'});
}

document.querySelector("#gethtml").onclick = getHtml;

function addSnippet(fname) {
    //TODO: add indent
    editor.find("end form");
    editor.navigateLineStart();
    editor.insert("\n");
    editor.navigateUp();
    snippetManager.insertSnippet(editor,snippetManager.getSnippetByName(fname,editor).content);
}

function debounce(func, wait, immediate) {
    var timeout;
    return function() {
        var context = this, args = arguments;
        var later = function() {
            timeout = null;
            if (!immediate) func.apply(context, args);
        };
        var callNow = immediate && !timeout;
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
        if (callNow) func.apply(context, args);
    };
}
let formHtml;
function parseForm() {
    let parser = new nearley.Parser(grammar);
    let formText = editor.getValue();
    parser.feed(formText);
    let formJson = parser.results[0];
    console.log(formJson);
//        jsonResultEditor.setValue(JSON.stringify(formJson, null,2));
    formHtml = builder(formJson);
    // setInnerHtml(document.querySelector("#form"), formHtml);
    document.querySelector("#form").innerHTML = formHtml;
    updateViewer("#form");
}
parseForm(); //first time

// function setInnerHtml (elm, html) {
//     elm.innerHTML = html;
//     Array.from(elm.querySelectorAll("script")).forEach( oldScript => {
//         const newScript = document.createElement("script");
//         Array.from(oldScript.attributes)
//             .forEach( attr => newScript.setAttribute(attr.name, attr.value) );
//         newScript.appendChild(document.createTextNode(oldScript.innerHTML));
//         oldScript.parentNode.replaceChild(newScript, oldScript);
//     });
// }


function escapeHtml(text) {
    var map = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#039;'
    };

    return text.replace(/[&<>"']/g, function(m) { return map[m]; });
}

function handleCopy() {
    let clipboard = new ClipboardJS('#copyToClipboard', {
        text: function(trigger) {
            return sharedHtml;
        }
    });
    clipboard.on('success', function(e) {
        console.info('Action:', e.action);
        console.info('Text:', e.text);
        console.info('Trigger:', e.trigger);

        e.clearSelection();
    });

    clipboard.on('error', function(e) {
        console.error('Action:', e.action);
        console.error('Trigger:', e.trigger);
    });
}



handleCopy();


//side panel trigger
(function(){
    // Slide In Panel - by CodyHouse.co
    var panelTriggers = document.getElementsByClassName('js-cd-panel-trigger');
    if( panelTriggers.length > 0 ) {
        for(var i = 0; i < panelTriggers.length; i++) {
            (function(i){
                var panelClass = 'js-cd-panel-'+panelTriggers[i].getAttribute('data-panel'),
                    panel = document.getElementsByClassName(panelClass)[0];
                // open panel when clicking on trigger btn
                panelTriggers[i].addEventListener('click', function(event){
                    event.preventDefault();
                    addClass(panel, 'cd-panel--is-visible');
                });
                //close panel when clicking on 'x' or outside the panel
                panel.addEventListener('click', function(event){
                    if( hasClass(event.target, 'js-cd-close') || hasClass(event.target, panelClass)) {
                        event.preventDefault();
                        removeClass(panel, 'cd-panel--is-visible');
                    }
                });
            })(i);
        }
    }

    //class manipulations - needed if classList is not supported
    //https://jaketrent.com/post/addremove-classes-raw-javascript/
    function hasClass(el, className) {
        if (el.classList) return el.classList.contains(className);
        else return !!el.className.match(new RegExp('(\\s|^)' + className + '(\\s|$)'));
    }
    function addClass(el, className) {
        if (el.classList) el.classList.add(className);
        else if (!hasClass(el, className)) el.className += " " + className;
    }
    function removeClass(el, className) {
        if (el.classList) el.classList.remove(className);
        else if (hasClass(el, className)) {
            var reg = new RegExp('(\\s|^)' + className + '(\\s|$)');
            el.className=el.className.replace(reg, ' ');
        }
    }
})();