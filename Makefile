CONFIG_PATH=${HOME}/.proglog/

.PHONY: init
init:
	mkdir -p ${CONFIG_PATH}

.PHONY: gencert
gencert:
		cfssl gencert \
			-initca test/ca-csr.json | cfssljson -bare ca
		cfssl gencert \
			-ca=ca.pem \
			-ca-key=ca-key.pem \
			-config=test/ca-config.json \
			-profile=server \
			test/server-csr.json | cfssljson -bare server
		cfssl gencert \
			-ca=ca.pem \
			-ca-key=
		mv *.pem *.csr ${CONFIG_PATH}

.PHONY:compile
compile:
	protoc api/v1/log.proto --proto_path=. --go_out=. --go_opt=paths=source_relative ./api/v1/log.proto --go-grpc_out=. --go-grpc_opt=paths=source_relative ./api/v1/log.proto

.PHONY:test
test:
	go test -race ./...