.EXPORT_ALL_VARIABLES:
gen:
	protoc -I api/ api.proto --go_out=plugins=grpc:internal
test:
	go test -v ./...