#!/bin/sh
grunt build
rm -rf ../assets/statics/
cp -r ./dist/template ../assets/
rm -rf ./dist/template
cp -r ./dist ../assets/statics/


