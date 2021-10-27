BIN_DIR		:= ./bin
CMD_DIR		:= ./cmd
WEB_DIR		:= ./web

SVC_TARGETS	:= $(addprefix $(BIN_DIR)/,$(patsubst $(CMD_DIR)/%/,%,$(dir $(wildcard $(CMD_DIR)/*/))))
SVC_SOURCES	:= $(shell find . -name "*.go")
WEB_TARGETS	:= $(WEB_DIR)/build
WEB_SOURCES	:= $(shell find "$(WEB_DIR)" -path "$(WEB_TARGETS)" -prune)

VERSION		?= dev


.PHONY: all run clean

# Compile all microservices and build frontend.
all: $(SVC_TARGETS) $(WEB_TARGETS)

# Compile a microservice.
$(SVC_TARGETS): $(BIN_DIR)/%: $(SVC_SOURCES)
	@mkdir -p $(@D)
	CGO_ENABLED=0 go build -o $@ -ldflags "-X main.name=$(@F) -X main.version=$(VERSION)" $(CMD_DIR)/$(@F)/main.go


$(WEB_TARGETS): $(WEB_SOURCES)
	cd $(WEB_DIR) && npm ci && npm run build

deploy:
	docker-compose up

clean:
	-@rm -rvf $(SVC_TARGETS)
	-@rm -rvf $(WEB_TARGETS)
