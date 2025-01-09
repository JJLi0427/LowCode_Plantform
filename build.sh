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

 cp -r pretools build/webtools/

fi

#cp -r schema-exec build/webtools/readme




