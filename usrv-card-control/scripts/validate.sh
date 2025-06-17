#!/bin/sh

while read package;
do
  echo "Validating $package package..."
  test="go test -coverprofile=coverage.out ./${package}"
  eval "$test"
  go tool cover -func=coverage.out | sed 's~\([^/]\{1,\}/\)\{3\}~~' | column -t | sed '$d' | sort -g -r -k 3
done <.package-list

echo "Running golangci-lint..."
golangci-lint run
npm run lint-md
