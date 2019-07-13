install:
	go get \
		github.com/golang/protobuf/protoc-gen-go \
		github.com/jteeuwen/go-bindata/go-bindata \
		github.com/golang/mock/mockgen

generate:
	go-bindata -pkg migrations -ignore bindata -prefix ./migrations/ -o migrations/bindata.go ./migrations