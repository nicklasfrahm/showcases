BIN_DIR		:= ./bin
CMD_DIR		:= ./cmd

SVC_TARGETS	:= $(addprefix $(BIN_DIR)/,$(patsubst $(CMD_DIR)/%/,%,$(dir $(wildcard $(CMD_DIR)/*/))))
SVC_SOURCES	:= $(shell find . -name "*.go")

VERSION		?= dev
PROFILES	?= audit,status,gateway-http,mail


.PHONY: all up clean

# Compile all microservices.
all: $(SVC_TARGETS)

# Compile a microservice.
$(SVC_TARGETS): $(BIN_DIR)/%: $(SVC_SOURCES)
	@mkdir -p $(@D)
	CGO_ENABLED=0 go build -o $@ -ldflags "-X main.name=$(@F) -X main.version=$(VERSION)" $(CMD_DIR)/$(@F)

# Using buildkit significantly enhances the build speed.
up:
	COMPOSE_PROFILES=nats,prometheus,grafana,$(PROFILES) COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 \
	docker-compose -f deployments/docker-compose.yml --env-file .env \
	up --build --remove-orphans

clean:
	-@rm -rvf $(SVC_TARGETS)
	-@rm -rvf $(WEB_TARGETS)
