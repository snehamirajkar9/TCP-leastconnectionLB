BINARY_NAME=bin/server
DOCKER_IMAGE=leastconn-lb

.PHONY: test test-race simulate-upstreams stop-upstreams build run docker-build clean
test: 			## Run the tests.
	@go test -cover ./...

test-race: 		## Run the tests with race detector.
	@go test -race ./...

simulate-upstreams: ## Start the simulated upstreams.
	@cd test && docker-compose up

stop-upstreams: 	## Stop the simulated upstreams.
	@cd test && docker-compose down

build: 			## Build the project.
	@go build -o ${BINARY_NAME} main.go

run: 			## Run the project.
	@go run main.go

docker-build: 	## Build the docker image.
	@docker build -t ${DOCKER_IMAGE} -f Dockerfile .

docker-run: docker-build 	## Run the docker image.
	@docker run -p 9443:9443 ${DOCKER_IMAGE}

docker-compose-build: ## Builds the docker-compose image
	@docker-compose build

docker-compose: docker-compose-build ## Run the docker-compose image
	@docker-compose up

docker-compose-down: ## Stop the docker-compose image
	@docker-compose down

clean: 		## Clean the project.
	@go clean
	@rm -f ${BINARY_NAME}
	@docker rmi -f ${DOCKER_IMAGE}

help:           ## Show this help.
	@egrep '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'


