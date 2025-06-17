#!/bin/sh

# Create a list of packages on a single string separated by spaces.
# E.g. ./foo ./bar
packages=""
while read package;
do
  packages+="./$package "
done <.package-list

echo "Running coverage on $packages"
coverage="go test -coverprofile=coverage.out $packages"
eval "$coverage"
go tool cover -html=coverage.out
