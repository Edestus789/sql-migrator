FROM golang:1.24.1 as build

ENV BIN_FILE /bin/app
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

CMD go test -tags=integration ./...
