## Introduction

Minimal Go framework for microservice applications running in [AWS](https://aws.amazon.com) supporting:

- Multiple HTTP endpoints on different ports
- [AWS SQS](https://aws.amazon.com/sqs/) message processing
- [Prometheus](https://prometheus.io) metrics endpoint

## Build & Test

Build:

```go
go build ./...
```

Test:

```go
go test
```

## Getting Started

See [cmd/demo/main.go](cmd/demo/main.go) for an example app.