# Sidecar source code

Source code and Dockerfile of the injected sidecar container.

```
1. Build and tag docker image
cd utscapstone/docker/sidecar
docker build . -t kevinygu/capstone-sidecar:0.0.1
docker push kevinygu/capstone-sidecar:0.0.1

# only for local testing
docker build . -t sidecar
docker run -it --rm --net host -v ${HOME}/.kube/:/root/.kube/ -v ${PWD}:/app --cap-add=NET_ADMIN --cap-add=NET_RAW sidecar sh
```