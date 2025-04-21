#!/bin/bash

if [ "$#" -lt 1 ]; then
  echo "Usage: $0 <tag>" 
  exit 1
fi

VERSION="$1"

echo "building read"

cd read-files

docker build . -t read-files:$VERSION 

cd ..

echo "building write"

cd write-files

docker build write-files/Dockerfile -t write-files:$VERSION

exit 0
