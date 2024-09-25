#!/bin/bash -ex

cd tools

tools=$(go list -m -f '{{if not (or .Indirect .Main)}}{{.Path}}@{{.Version}}{{end}}' all)

for tool in $tools; do
    go install "$tool"
done
