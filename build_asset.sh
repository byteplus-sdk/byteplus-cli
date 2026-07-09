#!/bin/bash
url=$1
urlBranch=$2

if [ "$url" == "" ]
then
  echo 'Please set metadata repo url'
  exit
fi




#clean git cache before build
rm -rf byteplus-sdk-metadata
rm -rf .git/modules/byteplus-sdk-metadata
git config --local --unset submodule.byteplus-sdk-metadata.url
git config --local --unset submodule.byteplus-sdk-metadata.active
git rm --cached byteplus-sdk-metadata

rm -rf .gitmodules
touch .gitmodules

git submodule add "$url" byteplus-sdk-metadata
if [ "$urlBranch" != "" ]
then
 cd byteplus-sdk-metadata
 git checkout -b "$urlBranch" origin/"$urlBranch"
 cd ..
fi
if ! go run ./scripts/generate_explorer_descriptions.go --metadata-dir byteplus-sdk-metadata/metadata --out byteplus-sdk-metadata/explorer_descriptions/descriptions.json
then
  echo "skip explorer descriptions generation"
  mkdir -p byteplus-sdk-metadata/explorer_descriptions
  printf '{}\n' > byteplus-sdk-metadata/explorer_descriptions/descriptions.json
fi

go-bindata -pkg asset  -o asset/asset.go byteplus-sdk-metadata/metadata/... byteplus-sdk-metadata/explorer_descriptions/...
go-bindata -pkg typeset  -o typeset/typeset.go byteplus-sdk-metadata/metatype/...
go-bindata -pkg structset  -o structset/structset.go byteplus-sdk-metadata/structure/...


#clean git cache after build
rm -rf byteplus-sdk-metadata
rm -rf .git/modules/byteplus-sdk-metadata
git config --local --unset submodule.byteplus-sdk-metadata.url
git config --local --unset submodule.byteplus-sdk-metadata.active
git rm --cached byteplus-sdk-metadata

rm -rf .gitmodules

