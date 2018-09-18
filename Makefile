PROJECT_NAME := oidctl
REPO_PATH := github.com/jobteaser/oidctl
VENDOR_DIR := ./vendor

GIT_COMMIT := $(shell \
	git rev-parse HEAD)
GIT_STATE := $(shell \
	[ `git status --porcelain 2>/dev/null | wc -l` -eq 0 ] && echo "clean" || echo "dirty")
BUILD_DATE := $(shell \
	date -u +%FT%TZ)

GO_BIN ?= go
LDFLAGS := -ldflags "-w -s -X main.GitCommit=${GIT_COMMIT} -X main.GitState=${GIT_STATE} -X main.BuildDate=${BUILD_DATE}"

GO_TEST_DIRS := $(shell \
	find . -name "*_test.go" -not -path "${VENDOR_DIR}/*" | \
	xargs -I {} dirname {}  | \
	uniq)

all: show

build:
	GOARCH=amd64 CGO_ENABLED=0 ${GO_BIN} build ${LDFLAGS} -o ${PROJECT_NAME} cmd/oidctl/main.go

fmt: ${GO_SRC_DIRS}
	@for dir in $^; do \
		pushd ./$$dir > /dev/null ; \
		${GO_BIN} fmt ; \
		popd > /dev/null ; \
	done;

test: ${GO_TEST_DIRS}
	@for dir in $^; do \
		${GO_BIN} test ${REPO_PATH}/$$dir; \
	done;

clean:
	rm ${PROJECT_NAME}

info:
	@echo "PROJECT_NAME     = ${PROJECT_NAME}"
	@echo "COMMIT           = ${GIT_COMMIT}"
	@echo "BUILD_DATE       = ${BUILD_DATE}"
	@echo "GO_BIN           = ${GO_BIN}"
	@echo "LDFLAGS          = ${LDFLAGS}"
	@echo "REPO_PATH        = ${REPO_PATH}"
	@echo "SRC              = ${GO_SRC_DIRS}"
	@echo "TEST             = ${GO_TEST_DIRS}"

.PHONY: all build fmt test deploy clean info
