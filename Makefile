.PHONY: build docker serve

default:

build:
	go build -o k8s-oidc-helper .

docker:
	docker build -t docker.k8s.jobteaser.net/coretech/k8s-oidc-helper .

serve:
	go run .