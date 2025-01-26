//need before libs
//<script src="assets/js/sform/nearley-2.16.0.js"></script>
//<script src="assets/js/sform/grammar.js"></script>
//<script src="assets/js/sform/builder.js"></script>
function sformdispatchDataConfig(queryselector) {
    window.dispatchData = function(payload) {
        let parser = new nearley.Parser(grammar);
        parser.feed(payload);
        let formJson = parser.results[0];
        //console.log(formJson);
        //jsonResultEditor.setValue(JSON.stringify(formJson, null,2));
        formHtml = builder(formJson);
        // setInnerHtml(document.querySelector("#form"), formHtml);
        document.querySelector(queryselector).innerHTML = formHtml;
    }

    if (typeof Vue != 'undefined') {
      const MyPlugin = {
        install(Vue, options) {
          Vue.prototype.dispatchData = window.dispatchData
          Vue.dispatchData = Vue.prototype.dispatchData
        },
      }
      Vue.use(MyPlugin)
    }
}

function parseHtmlElement(nodestr) {
  return new DOMParser().parseFromString(nodestr, 'text/html').body.childNodes[0];
}

function parseUrlName(url) {
 return url.substring(url.lastIndexOf('/')+1);
}

//need libs
//<script src="https://cdn.staticfile.org/vue/2.2.2/vue.min.js"></script>
function vuedispatchDataConfig(queryselector, eventName) {
    window.dispatchData = function(payload){
        const dispatcher = document.querySelector(queryselector);
        dispatcher.dispatchEvent(new CustomEvent(eventName, { bubbles: true, detail: payload }))
    }

    if (typeof Vue != 'undefined'){
      const MyPlugin = {
        install(Vue, options) {
          Vue.prototype.dispatchData = window.dispatchData
          Vue.dispatchData = Vue.prototype.dispatchData
        },
      }
      Vue.use(MyPlugin)
    }
}

var _internal_submit_isloading = false;
function sformOnSubmitConfig(formqueryselector, target) {
    const form = document.querySelector(formqueryselector)
    form.addEventListener('submit', (event) => {
          if (_internal_submit_isloading){
            return;
          }

          if (typeof dispatchData == 'undefined'){
            return
          }
          event.preventDefault();

          let form_data = new FormData(form)
          if (typeof ValidateFormData == 'function' && !ValidateFormData(form_data)) {
            return
          }
          if (typeof ReWriteFormData == 'function') {
              form_data = ReWriteFormData(form_data)
          }

          var submitLable = document.querySelector(formqueryselector+" label.formlabelsubmit");
          if(submitLable == null) {
            submitLable = document.querySelector(formqueryselector+" input[type='submit']");
            if(submitLable){
              submitLable = submitLable.parentElement
            }
          }

          if (submitLable){
            var loading = document.createElement("div");
            loading.className = "loadingspin";
            submitLable.appendChild(loading)
            _internal_submit_isloading = true
          }

          if (target == 'nobackserver') {
            let object = {};
            let form_data = new FormData(form)
            form_data.forEach(function(value, key) {
                object[key] = value;
            });
            dispatchData(object);
            if (submitLable){
              submitLable.removeChild(loading);
              _internal_submit_isloading = false;
            }
            return
          }

          const xhr =  new XMLHttpRequest();
          xhr.onload = function(){
            if (xhr.status != 200) {
              dispatchData({xhrstatus: xhr.status})
              if (submitLable){
                submitLable.removeChild(loading);
                _internal_submit_isloading = false;
              }
            } else {
              try {  
                  const data = JSON.parse(xhr.responseText);
                  dispatchData(data);
                  if (submitLable){
                    submitLable.removeChild(loading);
                    _internal_submit_isloading = false;
                  }
              } catch (e) {
                  dispatchData(xhr.responseText)
                  if (submitLable){
                    submitLable.removeChild(loading);
                    _internal_submit_isloading = false;
                  }
              }
            }
          }
          xhr.onerror = function(e){
              dispatchData({xhrstatus: 500})
              if (submitLable){
                submitLable.removeChild(loading);
                _internal_submit_isloading = false;
              }
          }

          if (form.method == "post" || form.method == "POST"){
            xhr.open("POST", form.action);
            xhr.send(form_data);
          }else{
            let form_str = new URLSearchParams(form_data).toString();
            xhr.open("GET", form.action+'?'+ form_str);
            xhr.send()
          }
      })
}

document.ready = function (callback) {
  ///兼容FF,Google
  if (document.addEventListener) {
      document.addEventListener('DOMContentLoaded', function () {
          document.removeEventListener('DOMContentLoaded', arguments.callee, false);
          callback();
      }, false)
  }
   //兼容IE
  else if (document.attachEvent) {
      document.attachEvent('onreadystatechange', function () {
            if (document.readyState == "complete") {
                      document.detachEvent("onreadystatechange", arguments.callee);
                      callback();
             }
      })
  }
  else if (document.lastChild == document.body) {
      callback();
  }
}

window.AsyncPostJsonData = function(url, payload, onsuccess, onerror) {
  let xhr = new XMLHttpRequest()
  xhr.onerror = function() {
    if (typeof onerror == 'function') {
        onerror({})
    } else if (typeof onsuccess == 'function') {
        onsuccess({})
    }
  }
  xhr.onabort=xhr.onerror

  xhr.onload = function() {
    if (typeof onsuccess == 'function') {
      try {
        onsuccess(JSON.parse(xhr.responseText))
      } catch (error) {
        xhr.onerror()
      }
    }
  }

  xhr.open('POST', url)
  xhr.send(payload)
}

//function ReWriteFormData(form_data) {
//  //$1...$n
//
//  if (form_data.has('key')) {
//    let fn_cmd_key = '(val) =>{}';
//    form_data.set('key', fn_cmd_key(form_data.get('key')))
//  }
//
//    return form_data;
//}

function AsyncPostDataRegisterUrl(url) {

  window.AsyncPostDataAutoUrl = function(payload, onsuccess, onerror) {
    let xhr = new XMLHttpRequest()
    xhr.onerror = function() {
      if (typeof onerror == 'function') {
          onerror({})
      } else if (typeof onsuccess == 'function') {
          onsuccess({})
      }
    }
    xhr.onabort=xhr.onerror

    xhr.onload = function() {
      if (typeof onsuccess == 'function') {
          onsuccess(xhr.responseText)
      }
    }

    xhr.open('POST', url)
    xhr.send(payload)
  }

}
