#!/bin/sh
echo "> Running as '$(id -u):$(id -g)' in '$(pwd)'"
echo "> sh -c $@"

sh -c "$@"

