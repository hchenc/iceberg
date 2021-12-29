FROM 364554757/centos:7.4.1708.go.1.16.3

WORKDIR /opt/devops

COPY ./cmd/controller-manager controller-manager

ENTRYPOINT ["/opt/devops/controller-manager"]