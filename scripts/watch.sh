#!/usr/bin/env bash

# continuously execute in foreground
while true; do
	_build/envelope -w ./demo/ -d -l debug
done
