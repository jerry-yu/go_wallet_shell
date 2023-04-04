#!/bin/sh
go build --ldflags '-extldflags "-Wl,--allow-multiple-definition"'
