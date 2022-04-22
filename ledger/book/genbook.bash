#!/usr/bin/bash

mkdir src-book

pushd src
for SRCFILE in $(find . -name "*.md")
do
	mdexec -template='```sh
{{.Output}}```' $SRCFILE > ../src-book/$SRCFILE
done
popd

mdbook build
rsync -a src/webshots html-book/
rsync -a src/consoleshots html-book/

rm -rf src-book
