.PHONY: build

path_to_add := $(addsuffix /bin,$(subst :,/bin:,$(GOPATH)))
export PATH := $(path_to_add):$(PATH)

SHELL	 := /usr/bin/env bash

GO       := GO111MODULE=on go
ifeq (${ENABLE_VENDOR}, 1)
GOVENDORFLAG := -mod=vendor
endif

GOBUILD := CGO_ENABLED=0 $(GO) build $(BUILD_FLAG) -trimpath $(GOVENDORFLAG)

build: tidy
	$(GOBUILD) -ldflags '$(LDFLAGS)' -o bin/gh ./main.go

tidy:
	$(GO) mod tidy
