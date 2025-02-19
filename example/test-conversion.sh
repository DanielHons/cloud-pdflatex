#!/bin/zsh


if [[ -z "$1" ]]; then
    echo "Usage: $0 <cloud-pdflatex URL>"
    exit 1
fi


curl -X POST "$1/convert" \
    -H "Content-Type: application/octet-stream" \
     --data-binary @example/generated.tar.gz --output result-from-template.pdf