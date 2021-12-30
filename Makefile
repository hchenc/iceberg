IMG ?= 364554757/devops:latest

install:
	kubectl apply -f deploy/deploy.yaml

uninstall:
	kubectl delete -f deploy/deploy.yaml

docker-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o ./bin/controller-manager ./cmd/controller-manager/controller-manager.go
	docker build . -t ${IMG} -f deploy/Dockerfile
	docker push ${IMG}
