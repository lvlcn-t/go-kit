.DEFAULT_GOAL := help
SHELL := /bin/bash

MODULES := $(shell go list -m | cut -d'/' -f 4- | sed 's/^/.\//')
VERSION := "v0.2.0"

.PHONY: help
help: ### Display this help
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?### "} /^[a-zA-Z_-]+:.*?### / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: lint
lint: ### Runs all pre-commit hooks on all modules
	@pre-commit run -a

.PHONY: mod-tidy
mod-tidy: ### Runs go mod tidy on all modules
	@for module in $(MODULES); do \
		cd $$module; \
		go mod tidy; \
		cd - > /dev/null; \
	done

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
	@for module in $(MODULES); do \
		tag=$$(basename $$module)/$(VERSION); \
		echo "Tagging $$module with $$tag"; \
		git tag $$tag; \
		git push origin $$tag; \
	done