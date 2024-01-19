# Default step to run
.DEFAULT_GOAL := help

docker: dbuild dstart ## Run all steps

dbuild: ## Build the application and the docker images
	CGO_ENABLED=0 go build -v -ldflags="-s -w" main.go
	docker compose build
	
dstart: ## Start the application in docker in foreground
	docker compose up

build: ## Build the application only
	CGO_ENABLED=0 go build -v -ldflags="-s -w" main.go

help: ## Print this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'