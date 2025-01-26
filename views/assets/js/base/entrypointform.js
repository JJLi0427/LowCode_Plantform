
function selectedinputtext_onchange(event, dataid, inputid){
  var inputelrel = document.getElementById(inputid);
  var el = document.getElementById(dataid);
  document.querySelectorAll('#'+dataid+ ' li').forEach(element => {
    if (element.textContent.indexOf(event.target.value) != -1){
      if (element.classList.contains('displaynone')){
        element.classList.remove('displaynone');
      }
    }else{
      if (!element.classList.contains('displaynone')){
        element.classList.add('displaynone');
      }
    }
  });
  inputelrel.value = event.target.value
}

function selectedinputtext_selectli(event, val, inputid){
  var el = document.getElementById(inputid+"-fk");
  var elrel = document.getElementById(inputid);

  elrel.value = val
  el.value = event.target.textContent;
  event.preventDefault();
}

function selectedinputtext_hidelist(dataid){
  var el = document.getElementById(dataid);
  if (!el.classList.contains('displaynone')){
    el.classList.add('displaynone');
  }
}

function selectedinputtext_expandlist(dataid) {
  var el = document.getElementById(dataid);
  if (el.classList.contains('displaynone')){
    el.classList.remove('displaynone');
  }
}

function selectedinputtext_mutationfn(inputid1, inputid2) {
    const firstInput = document.getElementById(inputid1); 
    const secondInput = document.getElementById(inputid2);
    const observer = new MutationObserver(()=>{ secondInput.disabled = firstInput.disabled; });
    observer.observe(firstInput, { attributes: true }); 
}

function isRegularExpression(str) {
  try {
    return str instanceof RegExp;
  } catch (e) {
    return false;
  }
}