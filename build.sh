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

EXENANME='webtools'
OS=$(systemOS)
if [[ ${OS} =~ 'WINDOWS' ]];then
EXENANME="${EXENANME}.exe"
fi

app=$1
apppath=$2

if [ ! -f views/assets/thirdparties/vuetify/vue.js ]; then
  curl https://cdn.jsdelivr.net/npm/vue@2.7.8/dist/vue.js -o views/assets/thirdparties/vuetify/vue.js
fi

if [ ! -f views/assets/thirdparties/vuetify/vuetify.js ]; then
  curl https://cdn.jsdelivr.net/npm/vuetify@2.6.8/dist/vuetify.js -o views/assets/thirdparties/vuetify/vuetify.js
fi

if [ ! -f views/assets/thirdparties/vuetify/vuetify.min.css ]; then
  curl https://cdn.jsdelivr.net/npm/vuetify@2.6.8/dist/vuetify.min.css -o views/assets/thirdparties/vuetify/vuetify.min.css
fi

if [ ${#app} -gt 2 ];then

 go build -o ${EXENANME} main.go

devapps="apps"
 if [ ${#apppath} -gt 3 ] && [ -d $apppath ];then
	devapps=$apppath
 fi

 ./${EXENANME} -debug  $app  -apppath $devapps

else

 [ -d build ] && rm -rf build

 mkdir -p build/webtools

 go build -o build/webtools/${EXENANME}  main.go

 cp -r views build/webtools/

 cp -r apps  build/webtools/

fi

#cp -r schema-exec build/webtools/readme




