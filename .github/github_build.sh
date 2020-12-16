#!/bin/sh

echo "=========print uname========="
uname -a
echo "=========print pwd =========="
pwd

echo "=========GOROOT=${GOROOT}========="
echo "=========go env========="
go env &&
echo "=========install coreutils jtool ========="
brew install coreutils
brew  install jtool

echo "=========which go========="
which go

echo "=========go get ========="
go get -d ./...
go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo
echo "=========go get done!========="

echo "=========start build========="
chmod +x ./build.sh
./build.sh
echo "=========go build done!========="

echo  "list files "
ls -lah
echo "=========find zips========="
find . -name "BaiduPCS-Go-*.zip"


echo "=========start upload========="
today=$(date +"%Y_%m_%d-%H_%M_%S")

nameSuff="_${today}${GITHUB_SHA}"

find . -name "BaiduPCS-Go-*.zip" -exec bash .github/upload-script.sh {} \;