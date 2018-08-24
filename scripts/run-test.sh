#! /usr/bin/env bash

export LOGXI="unalogger=OFF"

go test -v -race "$(go list ./... | grep -v /vendor/)"

