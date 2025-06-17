# usrv-card-control
This microservice request to registry, restore and block card of Mastercard and Visa.

## Pre-steps
If this is your first time using go and GoLand as an IDE you should check the following tutorial:
*   [Installing and configuring Go](https://kushki.atlassian.net/wiki/spaces/DT/pages/2068480362/Instalaci%2Bn%2By%2Bconfiguraci%2Bn%2Bde%2BGo)

### Commands
This commands will download our dependencies.

```shel
$ npm i
$ go mod download
$ go mod tidy
$ go mod vendor
```

### Deploy
Every deploy will trigger a webhook, which will deploy our code to AWS CodePipeline

QA - STG
If you want test your code deploy in QA-STG, create a PullRequest from your release branch to master.

PROD, UAT
If you want test your code deploy in PROD-UAT, Merge your PullRequest from your release to master.

## Running the tests
Use the following command to run your tests

Using the Makefile:

```shell
$ make test
$ make validate
$ make coverage
```

Using bash:

```shell
$ bash scripts/test.sh
$ bash scripts/validate.sh
$ bash scripts/coverage.sh
```

If you want to check which lines are not being covered you can use this command

```shell
$ make coverage
$ bash scripts/coverage.sh
```

### What you should never forget
You’ll start by editing this README file to learn how to edit a file in Bitbucket.

*   Read about CDK and Go to edit this project.
*   You must create a pull request to merge your code.
*   You must not commit at master.

Next, you’ll add a new file to this repository.

## Built With
*   [@kushki/cdk](https://bitbucket.org/kushki/kushki-cdk/src/master/) - Kushki CDK to deploy AWS resources
*   [Go](https://golang.org/) - The Go programming language
*   [Aws-Sdk-Go](https://github.com/aws/aws-sdk-go) - AWS SDK for go
*   [Aws-Lambda-Go](https://github.com/aws/aws-lambda-go/) - AWS package for lambdas
*   [Testify](https://github.com/stretchr/testify) - Library for testing
*   [Kushki Core](https://bitbucket.org/kushki/usrv-go-core/src/master/) - Kushki's library for microservices development
*   ❤️

## Acknowledgments

"Gofmt&#39;s style is no one&#39;s favorite, yet gofmt is everyone&#39;s favorite." - Rob Pike
