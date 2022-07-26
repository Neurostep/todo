# Simple Todo application

This repository contains a simple web application, representing as an HTTP API
service for a Simple TODO application.

Features:
* CRUD for TODO item
* Add a comment for the TODO
* Add a label for TODO

You can find an API spec in the [api/swagger.yaml](https://github.com/Neurostep/todo/blob/83a54c9984584ef3957ed74f88319e196fbb75ef/api/swagger.yaml)

## Run the application locally

To simply run this application, all you have to have installed is Docker engine with docker-compose toolset.
Once you have it installed, please follow steps:

1. Download this repo
2. run `docker-compose build`
3. run `docker-compose up -d`

The application will be exposed on port 19000. So, go to the [http://localhost:19001/api/v1/todos](http://localhost:19001/api/v1/todos) to check
If the application run correctly, you can start playing with the API. Use swaggerfile mentioned above as a reference.

## Deployment

The application delivered within staged Dockerfile. Thus we can easily use different strategies to deploy it.

#### Kubernetes

If we want to deploy the application using Kubernetes, we can simply build a docker image out of the Dockerfile
and put it into your favorite registry. Application is stateless, so we can schedule it easily using Kubernetes
Deployment object. We should create ConfigMap in order to provide configuration .yaml file
It is meant that before starting application there should be Postgres DB running. We should keep this in mind
and don't forget to provide the DSN to the application.

Application exposes 2 endpoints to check health/readiness: `/healthz` & `/readyz`

#### Run as a binary

If we want to deploy this app as a binary, we can do it using enormous possible ways. The repo provides Makefile.
By running `make build` the binary for current OS/arch will be compiled. `make build-release` will produce
linux-specific production ready binary, consult Makefile to learn more.

## Test

To run test, simply type a command `make test`.

## To improve

1. Test coverage
2. Provide k8s config files as a reference and a way to deploy
3. Extend API to include comments/labels into the TODOs response
4. Multi-user API, with authentication/authorization flow
...
