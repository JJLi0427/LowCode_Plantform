"use strict";

//need before libs
//<script src="assets/js/sform/nearley-2.16.0.js"></script>
//<script src="assets/js/sform/grammar.js"></script>
//<script src="assets/js/sform/builder.js"></script>
function sformdispatchDataConfig(queryselector) {
  window.dispatchData = function (payload) {
    var parser = new nearley.Parser(grammar);
    parser.feed(payload);
    var formJson = parser.results[0]; //console.log(formJson);
    //jsonResultEditor.setValue(JSON.stringify(formJson, null,2));

    formHtml = builder(formJson); // setInnerHtml(document.querySelector("#form"), formHtml);

    document.querySelector(queryselector).innerHTML = formHtml;
  };

  if (typeof Vue != 'undefined') {
    var MyPlugin = {
      install: function install(Vue, options) {
        Vue.prototype.dispatchData = window.dispatchData;
        Vue.dispatchData = Vue.prototype.dispatchData;
      }
    };
    Vue.use(MyPlugin);
  }
} //need libs
//<script src="https://cdn.staticfile.org/vue/2.2.2/vue.min.js"></script>


function vuedispatchDataConfig(queryselector, eventName) {
  window.dispatchData = function (payload) {
    var dispatcher = document.querySelector(queryselector);
    dispatcher.dispatchEvent(new CustomEvent(eventName, {
      bubbles: true,
      detail: payload
    }));
  };

  if (typeof Vue != 'undefined') {
    var MyPlugin = {
      install: function install(Vue, options) {
        Vue.prototype.dispatchData = window.dispatchData;
        Vue.dispatchData = Vue.prototype.dispatchData;
      }
    };
    Vue.use(MyPlugin);
  }
}

function sformOnSubmitConfig(formqueryselector, target) {
  var form = document.querySelector(formqueryselector);
  form.addEventListener('submit', function (event) {
    if (typeof dispatchData == 'undefined') {
      return;
    }

    if (target == 'nobackserver') {
      var object = {};

      var _form_data = new FormData(form);

      _form_data.forEach(function (value, key) {
        object[key] = value;
      });

      dispatchData(object);
      return;
    }

    var xhr = new XMLHttpRequest();

    xhr.onload = function () {
      if (xhr.status != 200) {
        dispatchData({
          xhrstatus: xhr.status
        });
      } else {
        try {
          var data = JSON.parse(xhr.responseText);
          dispatchData(data);
        } catch (e) {
          dispatchData(xhr.responseText);
        }
      }
    };

    xhr.onerror = function (e) {
      dispatchData({
        xhrstatus: 500
      });
    };

    var form_data = new FormData(form);

    if (form.method == "post" || form.method == "POST") {
      xhr.open("POST", form.action);
      xhr.send(form_data);
    } else {
      var form_str = new URLSearchParams(form_data).toString();
      xhr.open("GET", form.action + '?' + form_str);
      xhr.send();
    }

    event.preventDefault();
  });
}