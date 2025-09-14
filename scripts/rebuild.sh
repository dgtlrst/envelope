#!/usr/bin/env bash

while true; do
	go build -o _build/envelope cmd/main.go && pkill -f '_build/envelope'
	inotifywait -e attrib $(find . -name '*.go') || exit
done
