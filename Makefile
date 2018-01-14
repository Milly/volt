
NAME := volt
SRC := $(shell find . -type d -name 'vendor' -prune -o -type f -name '*.go' -print)
VERSION := $(shell sed -n -E 's/var voltVersion string = "([^"]+)"/\1/p' cmd/version.go)
RELEASE_LDFLAGS := -s -w -extldflags '-static'
RELEASE_OS := linux windows darwin
RELEASE_ARCH := amd64 386

DIST_DIR := dist
BIN_DIR := bin

all: $(BIN_DIR)/$(NAME)

$(BIN_DIR)/$(NAME): $(SRC)
	go build -o $(BIN_DIR)/$(NAME)

precompile:
	go build -a -i -o $(BIN_DIR)/$(NAME)
	rm $(BIN_DIR)/$(NAME)

install-dep:
	curl -L -o bin/dep https://github.com/golang/dep/releases/download/v0.3.2/dep-linux-amd64
	chmod +x bin/dep
	echo 'Installed dep v0.3.2 to bin/dep'

test:
	make
	go test -v -race -parallel 3 ./...

# Make static-linked binaries and tarballs
release: $(BIN_DIR)/$(NAME)
	rm -fr $(DIST_DIR)
	@for os in $(RELEASE_OS); do \
		for arch in $(RELEASE_ARCH); do \
			exe=$(DIST_DIR)/$(NAME)-$(VERSION)-$$os-$$arch; \
			if [ $$os = windows ]; then \
				exe=$$exe.exe; \
			fi; \
			echo "Creating $$exe ... (os=$$os, arch=$$arch)"; \
			GOOS=$$os GOARCH=$$arch go build -tags netgo -installsuffix netgo -ldflags "$(RELEASE_LDFLAGS)" -o $$exe; \
		done; \
	done

update-doc: all
	go run _scripts/update-readme.go README.md
	go run _scripts/update-cmdref.go CMDREF.md

.PHONY: all precompile install-dep test release update-doc
