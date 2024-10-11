.DEFAULT_GOAL := help
SHELL := /bin/bash

MODULES := $(shell go list -m | cut -d'/' -f 4- | sed 's/^/.\//')
VERSION := $(shell git describe --tags --abbrev=0 --match "v[0-9]*.[0-9]*.[0-9]*" 2>/dev/null)

.PHONY: help
help: ### Display this help
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?### "} /^[a-zA-Z_-]+:.*?### / {printf "  %-20s - %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: tidy
tidy: ### Runs go mod tidy on all modules
	@go work sync
	@for module in $(MODULES); do \
		cd $$module; \
		go mod tidy; \
		cd - > /dev/null; \
	done

.PHONY: update 
update: ### Updates all modules to the latest version
	@for module in $(MODULES); do \
		cd $$module; \
		go get -u; \
		cd - > /dev/null; \
	done

.PHONY: lint
lint: ### Runs linters on all modules
	@pre-commit run -a

.PHONY: example
example: ### Runs the example for [MODULE]
	@if [ -z "$(MODULE)" ]; then \
		echo "No module found"; \
		exit 1; \
	fi
	@cd example/$$(basename $(MODULE)); \
	go run ./simple; \
	cd - > /dev/null;

.PHONY: tag
tag: ### Tags the [MODULE] with [VERSION]
	@if [ -z "$(VERSION)" || -z "$(MODULE)" ]; then \
		echo "No version or module found"; \
		exit 1; \
	fi
	@read -p "Are you sure you want to tag $(MODULE) with $(VERSION)? [y/N] " -n 1 -r; \
	if [[ ! $$REPLY =~ ^[Yy]$$ ]]; then \
		echo ""; \
		echo "Aborting..."; \
		exit 1; \
	fi
	@tag=$$(basename $(MODULE))/$(VERSION); \
	echo "Tagging $(MODULE) with $$tag"; \
	git tag $$tag; \
	git push origin $$tag;

.PHONY: tag-all
tag-all: ### Tags all modules with [VERSION]
	@if [ -z "$(VERSION)" ]; then \
		echo "No version found"; \
		exit 1; \
	fi
	@read -p "Are you sure you want to tag all modules with $(VERSION)? [y/N] " -n 1 -r; \
	if [[ ! $$REPLY =~ ^[Yy]$$ ]]; then \
		echo ""; \
		echo "Aborting..."; \
		exit 1; \
	fi
	@echo ""
	@for module in $(MODULES); do \
		if [ $$(basename $$module) == "example" ]; then \
			continue; \
		fi; \
		tag=$$(basename $$module)/$(VERSION); \
		echo "Tagging $$module with $$tag"; \
		git tag $$tag; \
		git push origin $$tag; \
	done