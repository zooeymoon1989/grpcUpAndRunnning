proto_server:
	protoc -I./proto  --go-grpc_out=./my_guest_server/grpc --go_out=./my_guest_server/grpc  ./proto/my_guest.proto

proto_client:
	protoc -I./proto  --go-grpc_out=./my_guest_client/grpc --go_out=./my_guest_client/grpc  ./proto/my_guest.proto