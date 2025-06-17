compile = env GOOS=linux GOARCH=arm64 go build -o
upx_url = https://github.com/upx/upx/releases/download/v3.96
upx_file = upx-3.96-arm64_linux
percent = go tool cover -func=coverage.out | sed 's~\([^/]\{1,\}/\)\{3\}~~' | sed '$d' | sort -g -r -k 3
test_to_file = go test -coverprofile=coverage.out
current_path = $(shell pwd)
current_dir = $(shell basename "$(PWD)")
gopath = $(GOPATH)

build:
	rm -rf my-artifacts
	mkdir my-artifacts
	for f in ./cmd/*; \
	do ( \
		pathname=$$(basename $$f); \
		filename=$${pathname%.*}; \
		$(compile) bin/$$filename/bootstrap -tags lambda.norpc cmd/$${pathname}/$${pathname}.go; \
		cp .kushkirc bin/$$filename; \
		zip -j -r $$filename.zip bin/$$filename; \
		mv $$filename.zip my-artifacts; \
	); done

compress:
	chmod -R 777 bin/
	wget $(upx_url)/$(upx_file) && tar -C /usr/local -xf $(upx_file).tar.xz
	/usr/local/$(upx_file)/upx bin/*
	rm -rf /usr/local/$(upx_file)

test:
	go test -count=1 ./service/... ./gateway/... ./tools/... ./config/...

coverage:
	$(test_to_file) ./service/... ./gateway/... ./tools/... ./config/...
	go tool cover -html=coverage.out

validate:
	make imports
	make gofmt
	make lint
	$(test_to_file) ./service  && $(percent)
	$(test_to_file) ./gateway/...  && $(percent)
	$(test_to_file) ./tools  && $(percent)
	$(test_to_file) ./config/...  && $(percent)

lint:
	golangci-lint run --fast

percent:
	$(percent)

docs:
	rm -rf $(gopath)/src/godoc/src
	mkdir -p $(gopath)/src/godoc/src
	rsync -a $(current_path) $(gopath)/src/godoc/src --exclude "node_modules" --exclude ".git" --exclude "mocks"
	(cd $(gopath)/src/godoc; godoc -goroot .)
	cd $(current_path)

mock:
	mockery --dir ./service --output ./mocks/service --all
	mockery --dir ./gateway --all --output ./mocks/gateway
	mockery --dir ./vendor/bitbucket.org/kushki/usrv-go-core/logger --name="KushkiLogger" --output ./mocks/core
	mockery --dir ./vendor/bitbucket.org/kushki/usrv-go-core/gateway/dynamo --name="IDynamoGateway" --output ./mocks/core

mock-clear:
	find mocks -type f ! -name "*main.go" -delete

imports:
	find . -name \*.go ! -path "./mocks/*" ! -path "./node_modules/*" ! -path "./vendor/*" -exec goimports -w {} \;

gofmt:
	find . -name \*.go ! -path "./mocks/*" ! -path "./node_modules/*" ! -path "./vendor/*" -exec gofmt -s -w {} \;
