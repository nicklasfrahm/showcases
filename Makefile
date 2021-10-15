GOBUILD		:= go build
BIN_DIR		:= ./bin
CMD_DIR		:= ./cmd

TARGETS		:= $(addprefix $(BIN_DIR)/,$(patsubst $(CMD_DIR)/%/,%,$(dir $(wildcard $(CMD_DIR)/*/))))
SRCS		:= $(shell find . -name "*.go")

VERSION		?= dev


.PHONY: all run clean

# Compile all microservices.
all: $(TARGETS)

# Compile the given microservice. Use the option RUN=1 to
$(TARGETS): $(BIN_DIR)/%: $(SRCS)
	@mkdir -p $(@D)
	CGO_ENABLED=0 $(GOBUILD) -o $@ -ldflags "-X main.name=$(@F) -X main.version=$(VERSION)" $(CMD_DIR)/$(@F)/main.go

clean:
	-@rm -rvf $(BIN_DIR)/*
