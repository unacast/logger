#! /usr/bin/env bash

export LOGXI="unalogger=OFF"

go test -race "$(go list ./... | grep -v /vendor/)"

