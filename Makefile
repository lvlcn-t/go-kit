.DEFAULT_GOAL := lint
SHELL := /bin/bash

.PHONY: lint
lint:
	@pre-commit run --hook-stage pre-push -a
	@pre-commit run --hook-stage pre-commit -a
