# Webhook source code

This dir contains the source code and Dockerfile of webhook container.

To build the Docker container, follow the below steps.

```
1. Build and tag docker image
cd utscapstone/docker/webhook
docker build . -t kevinygu/capstone-webhook:0.0.1
docker push kevinygu/capstone-webhook:0.0.1

Define main module
go mod init webhook

Build binary
go build webhook

Local testing command
docker run -it --rm --net host -v ${HOME}/.kube/:/root/.kube/ -v ${PWD}:/app webhook sh
```