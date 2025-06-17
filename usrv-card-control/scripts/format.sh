#!/bin/sh

# Format code using gofmt and goimports.
echo "Running gofmt..."
find . -name \*.go ! -path "./.git/*" ! -path "./mocks/*" ! -path "./node_modules/*" ! -path "./vendor/*" ! -path "./proto/*" -exec gofmt -s -w {} \;
echo "Running goimports..."
find . -name \*.go ! -path "./.git/*" ! -path "./mocks/*" ! -path "./node_modules/*" ! -path "./vendor/*" ! -path "./proto/*" -exec goimports -w {} \;
