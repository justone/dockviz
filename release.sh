#!/bin/bash

ORG=justone
NAME=dockviz
ARCHS="darwin/amd64 linux/amd64 windows/amd64"

set -ex

if [[ ! $(type -P gox) ]]; then
    echo "Error: gox not found."
    echo "To fix: run 'go get github.com/mitchellh/gox', and/or add \$GOPATH/bin to \$PATH"
    exit 1
fi

if [[ -z $GITHUB_TOKEN ]]; then
    echo "Error: GITHUB_TOKEN not set."
    exit 1
fi

if [[ ! $(type -P github-release) ]]; then
    echo "Error: github-release not found."
    exit 1
fi

VER=$1

if [[ -z $VER ]]; then
    echo "Need to specify version."
    exit 1
fi

PRE_ARG=
if [[ $VER =~ pre ]]; then
    PRE_ARG="--pre-release"
fi

# git tag $VER

echo "Building $VER"
echo

rm -v ${NAME}* || true
gox -ldflags "-X main.version=$VER" -osarch="$ARCHS"

echo "* " > desc
echo "" >> desc

echo "\`\`\`" >> desc
echo "$ sha1sum ${NAME}_*" >> desc
sha1sum ${NAME}_* >> desc
echo "$ sha256sum ${NAME}_*" >> desc
sha256sum ${NAME}_* >> desc
echo "\`\`\`" >> desc

vi desc

git push --tags
sleep 2

cat desc | github-release release $PRE_ARG --user ${ORG} --repo ${NAME} --tag $VER --name $VER --description -
for file in ${NAME}_*; do
    github-release upload --user ${ORG} --repo ${NAME} --tag $VER --name $file --file $file
done
