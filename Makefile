.PHONY: build

path_to_add := $(addsuffix /bin,$(subst :,/bin:,$(GOPATH)))
export PATH := $(path_to_add):$(PATH)

SHELL	 := /usr/bin/env bash

GO       := GO111MODULE=on go
ifeq (${ENABLE_VENDOR}, 1)
GOVENDORFLAG := -mod=vendor
endif

ifeq (${TRIMPATH}, 1)
GOTRIMPATH := -trimpath
endif

GOBUILD := CGO_ENABLED=0 $(GO) build $(BUILD_FLAG) ${GOTRIMPATH} $(GOVENDORFLAG)

build: tidy
	$(GOBUILD) -ldflags '$(LDFLAGS)' -o bin/gh ./main.go

tidy:
	$(GO) mod tidy

run-daily-review: build
	./bin/gh -c config/cfg.toml review

run-weekly-review: build
	./bin/gh -c config/cfg.toml review weekly

run-monthly-review: build
	./bin/gh -c config/cfg.toml review monthly

run-daily-pkgs: build
	./bin/gh -c config/pkgs_cfg.toml pkgs

run-weekly-pkgs: build
	./bin/gh -c config/pkgs_cfg.toml pkgs weekly

run-monthly-pkgs: build
	./bin/gh -c config/pkgs_cfg.toml pkgs monthly
