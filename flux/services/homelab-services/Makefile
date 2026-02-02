.PHONY: help build-local build-ghcr push-local push-ghcr build-all-local build-all-ghcr push-all-local push-all-ghcr build-push-all

# Local registry
LOCAL_REGISTRY := localhost:5001

# Remote registry
GHCR_REGISTRY := ghcr.io
USER := brunolucena

# Service directories
MOBILE_API_DIR := mobile-api
REKORDBOX_CLOUD_DIR := rekordbox-cloud
SPOTIFY_P2P_DIR := spotify-p2p

# Local image names
MOBILE_API_LOCAL := $(LOCAL_REGISTRY)/homelab-services-mobile-api:latest
REKORDBOX_CLOUD_LOCAL := $(LOCAL_REGISTRY)/rekordbox-cloud-server:latest
SPOTIFY_P2P_LOCAL := $(LOCAL_REGISTRY)/spotify-p2p-server:latest

# GHCR image names
MOBILE_API_GHCR := $(GHCR_REGISTRY)/$(USER)/homelab-services-mobile-api:latest
REKORDBOX_CLOUD_GHCR := $(GHCR_REGISTRY)/$(USER)/rekordbox-cloud-server:latest
SPOTIFY_P2P_GHCR := $(GHCR_REGISTRY)/$(USER)/spotify-p2p-server:latest

help:
	@echo "Available targets:"
	@echo "  build-local            - Build all images to localhost:5001"
	@echo "  build-ghcr             - Build all images to ghcr.io"
	@echo "  push-local             - Push all images to localhost:5001"
	@echo "  push-ghcr              - Push all images to ghcr.io"
	@echo "  build-push-all          - Build to local, then push to ghcr.io"

# Build to local registry
build-local-mobile-api:
	@echo "Building $(MOBILE_API_LOCAL)..."
	docker build -t $(MOBILE_API_LOCAL) $(MOBILE_API_DIR)
	docker push $(MOBILE_API_LOCAL)

build-local-rekordbox:
	@echo "Building $(REKORDBOX_CLOUD_LOCAL)..."
	docker build -t $(REKORDBOX_CLOUD_LOCAL) $(REKORDBOX_CLOUD_DIR)
	docker push $(REKORDBOX_CLOUD_LOCAL)

build-local-spotify:
	@echo "Building $(SPOTIFY_P2P_LOCAL)..."
	docker build -t $(SPOTIFY_P2P_LOCAL) $(SPOTIFY_P2P_DIR)
	docker push $(SPOTIFY_P2P_LOCAL)

build-all-local: build-local-mobile-api build-local-rekordbox build-local-spotify
	@echo "✅ All images built and pushed to localhost:5001!"

# Build to GHCR
build-ghcr-mobile-api:
	@echo "Building $(MOBILE_API_GHCR)..."
	docker build -t $(MOBILE_API_GHCR) $(MOBILE_API_DIR)

build-ghcr-rekordbox:
	@echo "Building $(REKORDBOX_CLOUD_GHCR)..."
	docker build -t $(REKORDBOX_CLOUD_GHCR) $(REKORDBOX_CLOUD_DIR)

build-ghcr-spotify:
	@echo "Building $(SPOTIFY_P2P_GHCR)..."
	docker build -t $(SPOTIFY_P2P_GHCR) $(SPOTIFY_P2P_DIR)

build-all-ghcr: build-ghcr-mobile-api build-ghcr-rekordbox build-ghcr-spotify
	@echo "✅ All images built for ghcr.io!"

# Push to GHCR
push-ghcr-mobile-api: build-ghcr-mobile-api
	@echo "Pushing $(MOBILE_API_GHCR)..."
	docker push $(MOBILE_API_GHCR)

push-ghcr-rekordbox: build-ghcr-rekordbox
	@echo "Pushing $(REKORDBOX_CLOUD_GHCR)..."
	docker push $(REKORDBOX_CLOUD_GHCR)

push-ghcr-spotify: build-ghcr-spotify
	@echo "Pushing $(SPOTIFY_P2P_GHCR)..."
	docker push $(SPOTIFY_P2P_GHCR)

push-all-ghcr: build-all-ghcr
	@echo "Pushing all images to ghcr.io..."
	docker push $(MOBILE_API_GHCR)
	docker push $(REKORDBOX_CLOUD_GHCR)
	docker push $(SPOTIFY_P2P_GHCR)
	@echo "✅ All images pushed to ghcr.io!"

# Combined: build local, then build and push to GHCR
build-push-all: build-all-local push-all-ghcr
	@echo "✅ Build and push complete!"

