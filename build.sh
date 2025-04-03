#!/bin/bash
function systemOS() {
case "$OSTYPE" in
  solaris*) echo "SOLARIS" ;;
  darwin*)  echo "OSX" ;; 
  linux*)   echo "LINUX" ;;
  bsd*)     echo "BSD" ;;
  msys*)    echo "WINDOWS" ;;
  cygwin*)  echo "ALSO WINDOWS" ;;
  *)        echo "unknown: $OSTYPE" ;;
esac
}

EXENANME='lowcode_plantform'
OS=$(systemOS)
if [[ ${OS} =~ 'WINDOWS' ]];then
EXENANME="${EXENANME}.exe"
fi

app=$1
apppath=$2


mkdir -p views/assets/thirdparties/vuetify
mkdir -p views/assets/css/sform/fontawesome
mkdir -p views/assets/css/sform/webfonts

dependencies=(
  "views/assets/thirdparties/vuetify/vue.js|https://cdn.jsdelivr.net/npm/vue@2.7.8/dist/vue.js"
  "views/assets/thirdparties/vuetify/vuetify.js|https://cdn.jsdelivr.net/npm/vuetify@2.6.8/dist/vuetify.js"
  "views/assets/thirdparties/vuetify/vuetify.min.css|https://cdn.jsdelivr.net/npm/vuetify@2.6.8/dist/vuetify.min.css"
  "views/assets/thirdparties/vuetify/materialdesignicons.min.css|https://cdnjs.cloudflare.com/ajax/libs/MaterialDesign-Webfont/5.3.45/css/materialdesignicons.min.css"
  "views/assets/thirdparties/vuetify/materialdesignicons-webfont.eot|https://cdnjs.cloudflare.com/ajax/libs/MaterialDesign-Webfont/5.3.45/fonts/materialdesignicons-webfont.eot"
  "views/assets/thirdparties/vuetify/materialdesignicons-webfont.ttf|https://cdnjs.cloudflare.com/ajax/libs/MaterialDesign-Webfont/5.3.45/fonts/materialdesignicons-webfont.ttf"
  "views/assets/thirdparties/vuetify/materialdesignicons-webfont.woff|https://cdnjs.cloudflare.com/ajax/libs/MaterialDesign-Webfont/5.3.45/fonts/materialdesignicons-webfont.woff"
  "views/assets/thirdparties/vuetify/materialdesignicons-webfont.woff2|https://cdnjs.cloudflare.com/ajax/libs/MaterialDesign-Webfont/5.3.45/fonts/materialdesignicons-webfont.woff2"
  "views/assets/css/sform/fontawesome/brands.css|https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.8.2/css/brands.css"
  "views/assets/css/sform/fontawesome/fontawesome.css|https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.8.2/css/fontawesome.css"
  "views/assets/css/sform/fontawesome/solid.css|https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.8.2/css/solid.css"
  "views/assets/css/sform/webfonts/fa-brands-400.eot|https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.8.2/webfonts/fa-brands-400.eot"
  "views/assets/css/sform/webfonts/fa-brands-400.svg|https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.8.2/webfonts/fa-brands-400.svg"
  "views/assets/css/sform/webfonts/fa-brands-400.ttf|https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.8.2/webfonts/fa-brands-400.ttf"
  "views/assets/css/sform/webfonts/fa-brands-400.woff|https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.8.2/webfonts/fa-brands-400.woff"
  "views/assets/css/sform/webfonts/fa-brands-400.woff2|https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.8.2/webfonts/fa-brands-400.woff2"
  "views/assets/css/sform/webfonts/fa-solid-900.eot|https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.8.2/webfonts/fa-solid-900.eot"
  "views/assets/css/sform/webfonts/fa-solid-900.svg|https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.8.2/webfonts/fa-solid-900.svg"
  "views/assets/css/sform/webfonts/fa-solid-900.ttf|https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.8.2/webfonts/fa-solid-900.ttf"
  "views/assets/css/sform/webfonts/fa-solid-900.woff|https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.8.2/webfonts/fa-solid-900.woff"
  "views/assets/css/sform/webfonts/fa-solid-900.woff2|https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.8.2/webfonts/fa-solid-900.woff2"
)

for item in "${dependencies[@]}"; do
  file="${item%%|*}"
  url="${item##*|}"
  if [ ! -f "$file" ]; then
    curl -fSL "$url" -o "$file" || { 
      echo "Failed to download"; 
      exit 1; 
    }
  fi
done


if [ ${#app} -gt 2 ];then

 go build -o ${EXENANME} main.go

devapps="apps"
 if [ ${#apppath} -gt 3 ] && [ -d $apppath ];then
	devapps=$apppath
 fi

 ./${EXENANME} -debug  $app  -apppath $devapps

else

 [ -d build ] && rm -rf build

 mkdir -p build/

 go build -o build/${EXENANME}  main.go

 cp -r views build/

 cp -r apps  build/

fi