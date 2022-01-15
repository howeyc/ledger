#!/usr/bin/bash

git worktree add ./html-book gh-pages
mdbook build
rm -rf ./html-book/*
pushd ./html-book
git add -A
git commit -m 'deploy new book'
git push origin gh-pages
popd -
