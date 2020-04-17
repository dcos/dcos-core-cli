FROM golang:1.12

RUN curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin v1.21.0
RUN go get -u github.com/jstemmer/go-junit-report
