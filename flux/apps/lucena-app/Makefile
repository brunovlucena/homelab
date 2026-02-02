.PHONY: build push deploy

IMAGE_NAME = lucena-app
IMAGE_TAG = latest
REGISTRY = localhost:5001
FULL_IMAGE = $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

build:
	@echo "Building $(FULL_IMAGE)..."
	docker build -t $(FULL_IMAGE) .

push:
	@echo "Pushing $(FULL_IMAGE)..."
	docker push $(FULL_IMAGE)

deploy:
	@echo "Deploying to Kubernetes..."
	kubectl apply -k k8s/kustomize/base/

build-push: build push

all: build push deploy

