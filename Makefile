install-and-update-dependencies:
	go get ./...
	go get -u
	go mod tidy

generate-protos:
	cd protos && rm -f rss_grpc.pb.go rss.pb.go
	protoc --go_out=. --go-grpc_out=. protos/rss.proto

build: generate-protos
	rm -f rss-grpc
	go build -o rss-grpc *.go

build-and-run: build
	./rss-grpc

test:
	go test -v ./...
