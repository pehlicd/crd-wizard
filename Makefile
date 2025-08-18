# Color Definitions
NO_COLOR      = \033[0m
OK_COLOR      = \033[32;01m
ERROR_COLOR   = \033[31;01m
WARN_COLOR    = \033[33;01m

# Directories
APP_NAME      = crd-wizard
BIN_DIR       = ./bin
GO_BUILD      = $(BIN_DIR)/$(APP_NAME)

# Versioning and Build Information
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
COMMIT_SHA := $(shell git rev-parse HEAD)

# Define the name and tag for your Docker image.
IMAGE_NAME ?= $(APP_NAME)
IMAGE_TAG ?= $(VERSION)

# Main Targets
.PHONY: run serve run-ui build-ui build-ui-and-embed build-backend fmt docker-build create-cluster delete-cluster deploy-ingress-nginx clean

## Run the application in serve mode
run:
	@echo "$(OK_COLOR)==> Running the application...$(NO_COLOR)"
	go run . web

## Serve the application (build UI and backend first)
serve: build-ui-and-embed build-backend
	@echo "$(OK_COLOR)==> Starting the application...$(NO_COLOR)"
	go run . web

## Run UI in development mode
run-ui:
	@echo "$(OK_COLOR)==> Running UI in development mode...$(NO_COLOR)"
	cd ui && npm run dev

## Build the UI
build-ui:
	@echo "$(OK_COLOR)==> Building UI...$(NO_COLOR)"
	cd ui && npm run build

## Build UI, embed it into Go application, and clean up artifacts
build-ui-and-embed: build-ui
	@echo "$(OK_COLOR)==> Embedding UI files into Go application...$(NO_COLOR)"
	rm -rf ./internal/web/static/*
	mv ./ui/dist/* ./internal/web/static/
	@echo "$(OK_COLOR)==> Cleaning up UI build artifacts...$(NO_COLOR)"
	rm -rf ./ui/dist
	rm -rf ./ui/node_modules

## Run Terminal UI
run-tui:
	@echo "$(OK_COLOR)==> Running Terminal UI...$(NO_COLOR)"
	go run . tui


## Build the Go backend and place the binary in bin directory
build-backend:
	@echo "$(OK_COLOR)==> Building Go backend...$(NO_COLOR)"
	mkdir -p $(BIN_DIR)
	go build -o $(GO_BUILD)

## Format Go code and tidy modules
fmt:
	@echo "$(OK_COLOR)==> Formatting Go code and tidying modules...$(NO_COLOR)"
	go fmt ./...
	go mod tidy

## Build docker image
docker-build:
	@echo "$(OK_COLOR)==> Building Docker image...$(NO_COLOR)"
	@docker build \
      --build-arg VERSION=$(VERSION) \
      --build-arg BUILD_DATE=$(BUILD_DATE) \
      --build-arg COMMIT_SHA=$(COMMIT_SHA) \
      -t $(IMAGE_NAME):$(IMAGE_TAG) .

# Kubernetes Cluster Management
## Create a Kubernetes cluster using Kind
create-k8s-cluster:
	@echo "$(OK_COLOR)==> Creating Kubernetes cluster...$(NO_COLOR)"
	kind create cluster --config dev/kind.yaml

## Delete the Kubernetes cluster
delete-k8s-cluster:
	@echo "$(OK_COLOR)==> Deleting Kubernetes cluster...$(NO_COLOR)"
	kind delete cluster -n 'crd-wizard-dev'

## Deploy NGINX Ingress controller to the cluster
deploy-ingress-nginx:
	@echo "$(OK_COLOR)==> Deploying ingress-nginx...$(NO_COLOR)"
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
	@echo "$(OK_COLOR)==> Waiting for ingress-nginx to be ready...$(NO_COLOR)"
	kubectl wait --namespace ingress-nginx \
		--for=condition=ready pod \
		--selector=app.kubernetes.io/component=controller \
		--timeout=180s

# Cleanup
## Remove built binaries and cleanup
clean:
	@echo "$(OK_COLOR)==> Cleaning up build artifacts...$(NO_COLOR)"
	rm -rf $(BIN_DIR)
