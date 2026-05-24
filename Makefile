BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
APPNAME := byx

# do not override user values
ifeq (,$(VERSION))
  VERSION := $(shell git describe --exact-match 2>/dev/null)
  # if VERSION is empty, then populate it with branch name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

# Update the ldflags with the app, client & server names
ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=$(APPNAME) \
	-X github.com/cosmos/cosmos-sdk/version.AppName=$(APPNAME)d \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT)

BUILD_FLAGS := -ldflags '$(ldflags)'

##############
###  Test  ###
##############

test-unit:
	@echo Running unit tests...
	@go test -mod=readonly -v -timeout 30m ./...

test-race:
	@echo Running unit tests with race condition reporting...
	@go test -mod=readonly -v -race -timeout 30m ./...

test-cover:
	@echo Running unit tests and creating coverage report...
	@go test -mod=readonly -v -timeout 30m -coverprofile=$(COVER_FILE) -covermode=atomic ./...
	@go tool cover -html=$(COVER_FILE) -o $(COVER_HTML_FILE)
	@rm $(COVER_FILE)

bench:
	@echo Running unit tests with benchmarking...
	@go test -mod=readonly -v -timeout 30m -bench=. ./...

test: govet govulncheck test-unit

.PHONY: test test-unit test-race test-cover bench

#################
###  Install  ###
#################

all: install

install:
	@echo "--> ensure dependencies have not been modified"
	@go mod verify
	@echo "--> installing $(APPNAME)d"
	@go install $(BUILD_FLAGS) -mod=readonly ./cmd/$(APPNAME)d

.PHONY: all install

##################
###  Protobuf  ###
##################

# Use this target if you do not want to use Ignite for generating proto files

proto-deps:
	@echo "Installing proto deps"
	@echo "Proto deps present, run 'go tool' to see them"

proto-gen:
	@echo "Generating protobuf files..."
	@ignite generate proto-go --yes

.PHONY: proto-gen

#################
###  Linting  ###
#################

lint:
	@echo "--> Running linter"
	@go tool github.com/golangci/golangci-lint/cmd/golangci-lint run ./... --timeout 15m

lint-fix:
	@echo "--> Running linter and fixing issues"
	@go tool github.com/golangci/golangci-lint/cmd/golangci-lint run ./... --fix --timeout 15m

.PHONY: lint lint-fix

###################
### Development ###
###################

govet:
	@echo Running go vet...
	@go vet ./...

govulncheck:
	@echo Running govulncheck...
	@go tool golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

.PHONY: govet govulncheck

############################
### Webhook UBYX E2E     ###
############################

E2E_WEBHOOK_UBYX_DIR ?= .e2e/webhook-ubyx

preflight-webhook-ubyx:
	@mkdir -p $(E2E_WEBHOOK_UBYX_DIR)
	@bash -o pipefail -c 'BYX_REST="$${BYX_REST:-http://127.0.0.1:1317}" \
	BYX_RPC="$${BYX_RPC:-http://127.0.0.1:26657}" \
	bash scripts/preflight_webhook_ubyx.sh | tee $(E2E_WEBHOOK_UBYX_DIR)/preflight.log'

doctor-webhook-ubyx:
	@mkdir -p $(E2E_WEBHOOK_UBYX_DIR)
	@BYX_REST="$${BYX_REST:-http://127.0.0.1:1317}"; \
	BYX_RPC="$${BYX_RPC:-http://127.0.0.1:26657}"; \
	{ echo "=== BYX Webhook UBYX Doctor ==="; \
	echo "REST: $$BYX_REST"; \
	echo "RPC:  $$BYX_RPC"; \
	if curl -sf "$$BYX_REST/cosmos/base/tendermint/v1beta1/syncing" >/dev/null; then echo "REST status: OK"; else echo "REST status: FAIL (porta 1317 fechada ou processo chain indisponivel)"; fi; \
	if curl -sf "$$BYX_RPC/status" >/dev/null; then echo "RPC status: OK"; else echo "RPC status: FAIL (porta 26657 fechada ou processo chain indisponivel)"; fi; \
	CHAIN_ID=$$(curl -sf "$$BYX_RPC/status" 2>/dev/null | jq -r '.result.node_info.network // empty'); \
	if [ -n "$$CHAIN_ID" ]; then echo "Chain-id detectado: $$CHAIN_ID"; else echo "Chain-id detectado: (indisponivel)"; fi; \
	echo "Portas esperadas: chain REST 1317, chain RPC 26657, mock merchant 4000"; \
	echo ""; \
	echo "Comandos sugeridos:"; \
	echo "0) make stack-webhook-ubyx-up"; \
	echo "1) ignite chain serve --reset-once"; \
	echo "2) cd webhook-relay/mock-merchant && MERCHANT_WEBHOOK_SECRET=devsecret PORT=4000 npm start"; \
	echo "3) cd webhook-relay && REST_ENDPOINT=http://127.0.0.1:1317 LOJA_ID=1 MERCHANT_WEBHOOK_URL=http://127.0.0.1:4000/webhook MERCHANT_WEBHOOK_SECRET=devsecret STATE_PATH=./state.json npm start"; \
	echo "4) make e2e-webhook-ubyx"; \
	echo "5) logs em: $(E2E_WEBHOOK_UBYX_DIR)"; } | tee $(E2E_WEBHOOK_UBYX_DIR)/doctor.log

e2e-webhook-ubyx:
	@mkdir -p $(E2E_WEBHOOK_UBYX_DIR)
	@bash -o pipefail -c 'BYX_REST="$${BYX_REST:-http://127.0.0.1:1317}" \
	BYX_RPC="$${BYX_RPC:-http://127.0.0.1:26657}" \
	bash scripts/e2e_payments_webhook_ubyx.sh | tee $(E2E_WEBHOOK_UBYX_DIR)/e2e.log'

stack-webhook-ubyx-up:
	@mkdir -p $(E2E_WEBHOOK_UBYX_DIR)
	@BYX_REST="$${BYX_REST:-http://127.0.0.1:1317}" \
	BYX_RPC="$${BYX_RPC:-http://127.0.0.1:26657}" \
	MOCK_MERCHANT_URL="$${MOCK_MERCHANT_URL:-http://127.0.0.1:4000/webhook}" \
	MERCHANT_WEBHOOK_SECRET="$${MERCHANT_WEBHOOK_SECRET:-devsecret}" \
	MOCK_EVENTS_LOG_PATH="$${MOCK_EVENTS_LOG_PATH:-$(E2E_WEBHOOK_UBYX_DIR)/mock-events.jsonl}" \
	STATE_PATH="$${STATE_PATH:-$(PWD)/webhook-relay/state.json}" \
	STRICT_WEBHOOK="$${STRICT_WEBHOOK:-1}" \
	bash scripts/e2e_webhook_ubyx_stack_up.sh

stack-webhook-ubyx-down:
	@E2E_DIR="$(E2E_WEBHOOK_UBYX_DIR)" \
	bash scripts/e2e_webhook_ubyx_stack_down.sh

stack-webhook-ubyx-logs:
	@echo "Logs dir: $(E2E_WEBHOOK_UBYX_DIR)"
	@ls -1 $(E2E_WEBHOOK_UBYX_DIR) 2>/dev/null || true
	@echo ""
	@for f in chain.log mock-merchant.log webhook-relay.log preflight.log doctor.log e2e.log; do \
	  if [ -f "$(E2E_WEBHOOK_UBYX_DIR)/$$f" ]; then \
	    echo "=== $$f (tail -n 40) ==="; \
	    tail -n 40 "$(E2E_WEBHOOK_UBYX_DIR)/$$f"; \
	    echo ""; \
	  fi; \
	done

e2e-webhook-ubyx-full:
	@set +e; STATUS=0; \
	echo ">>> stack up"; \
	$(MAKE) stack-webhook-ubyx-up || STATUS=$$?; \
	if [ $$STATUS -eq 0 ]; then echo ">>> preflight"; $(MAKE) preflight-webhook-ubyx || STATUS=$$?; fi; \
	if [ $$STATUS -eq 0 ]; then echo ">>> doctor"; $(MAKE) doctor-webhook-ubyx || STATUS=$$?; fi; \
	if [ $$STATUS -eq 0 ]; then echo ">>> e2e"; $(MAKE) e2e-webhook-ubyx || STATUS=$$?; fi; \
	echo ">>> collect artifacts"; \
	E2E_DIR="$(E2E_WEBHOOK_UBYX_DIR)" \
	STATE_PATH="$${STATE_PATH:-$(PWD)/webhook-relay/state.json}" \
	MOCK_EVENTS_LOG_PATH="$${MOCK_EVENTS_LOG_PATH:-$(E2E_WEBHOOK_UBYX_DIR)/mock-events.jsonl}" \
	bash scripts/e2e_webhook_ubyx_collect_artifacts.sh; \
	echo ">>> stack down"; \
	$(MAKE) stack-webhook-ubyx-down; \
	exit $$STATUS

.PHONY: preflight-webhook-ubyx doctor-webhook-ubyx e2e-webhook-ubyx stack-webhook-ubyx-up stack-webhook-ubyx-down stack-webhook-ubyx-logs e2e-webhook-ubyx-full
