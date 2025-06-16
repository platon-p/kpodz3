FROM golang:1.23-alpine
# install protoc and proto-gen-go
RUN apk add --no-cache \
    protobuf \
    protobuf-dev
# install protoc-gen-go
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

WORKDIR /workspace

CMD ["protoc", "--go_out=proto", "--go_opt=module=github.com/platon-p/kpodz3/proto", "types.proto"]