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
	@bash -o pipefail -c 'BYX_CHAIN_MODE="$${BYX_CHAIN_MODE:-}" \
	BYX_REST="$${BYX_REST:-http://127.0.0.1:1317}" \
	BYX_RPC="$${BYX_RPC:-http://127.0.0.1:26657}" \
	BYXD_BIN="$${BYXD_BIN:-byxd}" \
	KEYRING_BACKEND="$${KEYRING_BACKEND:-test}" \
	MERCHANT_KEY="$${MERCHANT_KEY:-merchant}" \
	PAYER_KEY="$${PAYER_KEY:-payer}" \
	AMOUNT_UBYX="$${AMOUNT_UBYX:-500000}" \
	MIN_MERCHANT_BALANCE_UBYX="$${MIN_MERCHANT_BALANCE_UBYX:-1}" \
	MIN_PAYER_BALANCE_UBYX="$${MIN_PAYER_BALANCE_UBYX:-$${AMOUNT_UBYX:-500000}}" \
	bash scripts/preflight_webhook_ubyx.sh | tee $(E2E_WEBHOOK_UBYX_DIR)/preflight.log'

doctor-webhook-ubyx:
	@mkdir -p $(E2E_WEBHOOK_UBYX_DIR)
	@bash -o pipefail -c 'E2E_DIR="$(E2E_WEBHOOK_UBYX_DIR)" \
	BYX_CHAIN_MODE="$${BYX_CHAIN_MODE:-}" \
	BYX_REST="$${BYX_REST:-http://127.0.0.1:1317}" \
	BYX_RPC="$${BYX_RPC:-http://127.0.0.1:26657}" \
	bash scripts/doctor_webhook_ubyx.sh | tee $(E2E_WEBHOOK_UBYX_DIR)/doctor.log'

e2e-webhook-ubyx:
	@mkdir -p $(E2E_WEBHOOK_UBYX_DIR)
	@bash -o pipefail -c 'BYX_CHAIN_MODE="$${BYX_CHAIN_MODE:-}" \
	BYX_REST="$${BYX_REST:-http://127.0.0.1:1317}" \
	BYX_RPC="$${BYX_RPC:-http://127.0.0.1:26657}" \
	BYXD_BIN="$${BYXD_BIN:-byxd}" \
	STATE_PATH="$${STATE_PATH:-$(PWD)/$(E2E_WEBHOOK_UBYX_DIR)/state.json}" \
	bash scripts/e2e_payments_webhook_ubyx.sh | tee $(E2E_WEBHOOK_UBYX_DIR)/e2e.log'

stack-webhook-ubyx-up:
	@mkdir -p $(E2E_WEBHOOK_UBYX_DIR)
	@BYX_CHAIN_MODE="$${BYX_CHAIN_MODE:-}" \
	BYX_REST="$${BYX_REST:-http://127.0.0.1:1317}" \
	BYX_RPC="$${BYX_RPC:-http://127.0.0.1:26657}" \
	BYX_CHAIN_ID="$${BYX_CHAIN_ID:-}" \
	BYXD_BIN="$${BYXD_BIN:-byxd}" \
	BYX_HOME="$${BYX_HOME:-}" \
	BYX_CHAIN_START_CMD="$${BYX_CHAIN_START_CMD:-}" \
	MOCK_MERCHANT_URL="$${MOCK_MERCHANT_URL:-http://127.0.0.1:4000/webhook}" \
	MERCHANT_WEBHOOK_SECRET="$${MERCHANT_WEBHOOK_SECRET:-devsecret}" \
	MOCK_EVENTS_LOG_PATH="$${MOCK_EVENTS_LOG_PATH:-$(E2E_WEBHOOK_UBYX_DIR)/mock-events.jsonl}" \
	STATE_PATH="$${STATE_PATH:-$(PWD)/$(E2E_WEBHOOK_UBYX_DIR)/state.json}" \
	STRICT_WEBHOOK="$${STRICT_WEBHOOK:-1}" \
	bash scripts/e2e_webhook_ubyx_stack_up.sh

stack-webhook-ubyx-down:
	@E2E_DIR="$(E2E_WEBHOOK_UBYX_DIR)" \
	bash scripts/e2e_webhook_ubyx_stack_down.sh

stack-webhook-ubyx-logs:
	@echo "Logs dir: $(E2E_WEBHOOK_UBYX_DIR)"
	@ls -1 $(E2E_WEBHOOK_UBYX_DIR) 2>/dev/null || true
	@echo ""
	@for f in chain.log mock-merchant.log webhook-relay.log preflight.log doctor.log e2e.log chain_mode.txt env_summary.txt startup_command.txt failure_reason.txt; do \
	  if [ -f "$(E2E_WEBHOOK_UBYX_DIR)/$$f" ]; then \
	    echo "=== $$f (tail -n 40) ==="; \
	    tail -n 40 "$(E2E_WEBHOOK_UBYX_DIR)/$$f"; \
	    echo ""; \
	  fi; \
	done

e2e-webhook-ubyx-keys:
	@bash -o pipefail -c 'KEYRING_BACKEND="$${KEYRING_BACKEND:-test}" \
	MERCHANT_KEY="$${MERCHANT_KEY:-merchant}" \
	BYX_REST="$${BYX_REST:-http://127.0.0.1:1317}" \
	bash scripts/e2e_webhook_ubyx_keys_setup.sh'

e2e-webhook-ubyx-external:
	@BYX_CHAIN_MODE=external $(MAKE) e2e-webhook-ubyx-full

e2e-webhook-ubyx-byxd:
	@BYX_CHAIN_MODE=byxd $(MAKE) e2e-webhook-ubyx-full

e2e-webhook-ubyx-custom:
	@BYX_CHAIN_MODE=custom $(MAKE) e2e-webhook-ubyx-full

e2e-webhook-ubyx-full:
	@set +e; STATUS=0; \
	echo ">>> stack up"; \
	$(MAKE) stack-webhook-ubyx-up || STATUS=$$?; \
	if [ $$STATUS -eq 0 ]; then echo ">>> preflight"; $(MAKE) preflight-webhook-ubyx || STATUS=$$?; fi; \
	if [ $$STATUS -eq 0 ]; then echo ">>> doctor"; $(MAKE) doctor-webhook-ubyx || STATUS=$$?; fi; \
	if [ $$STATUS -eq 0 ]; then echo ">>> e2e"; $(MAKE) e2e-webhook-ubyx || STATUS=$$?; fi; \
	echo ">>> collect artifacts"; \
	E2E_DIR="$(E2E_WEBHOOK_UBYX_DIR)" \
	STATE_PATH="$${STATE_PATH:-$(PWD)/$(E2E_WEBHOOK_UBYX_DIR)/state.json}" \
	MOCK_EVENTS_LOG_PATH="$${MOCK_EVENTS_LOG_PATH:-$(E2E_WEBHOOK_UBYX_DIR)/mock-events.jsonl}" \
	bash scripts/e2e_webhook_ubyx_collect_artifacts.sh; \
	echo ">>> stack down"; \
	$(MAKE) stack-webhook-ubyx-down; \
	exit $$STATUS

.PHONY: preflight-webhook-ubyx doctor-webhook-ubyx e2e-webhook-ubyx stack-webhook-ubyx-up stack-webhook-ubyx-down stack-webhook-ubyx-logs e2e-webhook-ubyx-keys e2e-webhook-ubyx-full e2e-webhook-ubyx-external e2e-webhook-ubyx-byxd e2e-webhook-ubyx-custom
