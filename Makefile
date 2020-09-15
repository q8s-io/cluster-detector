# If you update this file, please follow
# https://suva.sh/posts/well-documented-makefiles

# Ensure Make is run with bash shell as some syntax below is bash-specific
SHELL:=/usr/bin/env bash

.DEFAULT_GOAL := help

# Code
VERSION := $(shell git rev-parse --short HEAD)
CODEPATH := $(shell go mod why | sed -n '2,2p')

# Tools
TOOLS_DIR := tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin

# Define Docker related variables. Releases should modify and double check these vars.
REGISTRY := uhub.service.ucloud.cn/devsecops
IMAGE := cluster-detector
CONTROLLER_IMG := $(REGISTRY)/$(IMAGE)
TAG := dev
ARCH := amd64

# Use GOPROXY environment variable if set
GOPROXY := $(shell go env GOPROXY)
ifeq ($(GOPROXY),)
GOPROXY := https://goproxy.cn
endif
export GOPROXY
# Active module mode, as we use go modules to manage dependencies
export GO111MODULE=on
# Lint
GOLANGCI_LINT := $(TOOLS_BIN_DIR)/golangci-lint

.PHONY: server

run:
	GOPROXY=https://goproxy.cn GO111MODULE=on go run detector.go -conf "./configs/pro.toml"


server:
	@echo "version: $(VERSION)"
	docker build --no-cache --build-arg CODEPATH=$(CODEPATH) -t $(REGISTRY)/$(IMAGE):$(VERSION) -f Dockerfile .
	docker push $(REGISTRY)/$(IMAGE):$(VERSION)
