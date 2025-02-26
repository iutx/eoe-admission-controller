# build info
PROJ_PATH := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
GOARCH ?= $(shell go env GOARCH)
GOOS ?= $(shell go env GOOS)
VERSION = $(shell cat VERSION)
GO_VERSION := $(shell go version)
GO_SHORT_VERSION := $(shell go version | awk '{print $$3}')
BUILD_TIME := $(shell date "+%Y-%m-%d %H:%M:%S")
COMMIT_ID := $(shell git rev-parse HEAD 2>/dev/null)
IMAGE_TAG := ${VERSION}-$(shell date '+%Y%m%d%H%M%S')-$(shell git rev-parse --short HEAD)
IMAGE_REPO := registry.erda.cloud/erda
BUILD_IMAGE := ${IMAGE_REPO}/eoe:${IMAGE_TAG}
LATEST_IMAGE := ${IMAGE_REPO}/eoe

build-image:
	@echo Start build image: ${BUILD_IMAGE}
	@docker build -f Dockerfile . -t ${BUILD_IMAGE}
	@docker tag ${BUILD_IMAGE} ${LATEST_IMAGE}

build-push-image: build-image
	@echo Start push image: ${BUILD_IMAGE}
	@docker push ${BUILD_IMAGE}
	@docker push ${LATEST_IMAGE}
