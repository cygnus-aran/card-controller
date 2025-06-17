#!/bin/sh

# Create a list of packages on a single string separated by spaces.
# E.g. ./foo ./bar
packages=""
while read package;
do
  packages+="./$package "
done <.package-list

echo "Running unit tests on $packages"
test="go test -count=2 $packages"
eval "$test"
