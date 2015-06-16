#!/bin/sh

make || exit $?
make gotest
