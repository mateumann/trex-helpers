
unexport GOBIN
GO           ?= go
GOFMT        ?= $(GO)fmt
FIRST_GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
GOLANGCI_LINT :=
GOLANGCI_LINT_OPTS ?=
ifeq ($(GOHOSTOS),$(filter $(GOHOSTOS),linux darwin))
ifeq ($(GOHOSTARCH),$(filter $(GOHOSTARCH),amd64 i386))
	GOLANGCI_LINT := $(shell which golangci-lint)
endif
endif
.PHONY: common-all
common-all: precheck style check_license lint unused build test

.PHONY: common-lint
common-lint: #$(GOLANGCI_LINT)
ifdef GOLANGCI_LINT
	@echo ">> running golangci-lint"
ifdef GO111MODULE
# 'go list' needs to be executed before staticcheck to prepopulate the modules cache.
# Otherwise staticcheck might fail randomly for some reason not yet explained.
	GO111MODULE=$(GO111MODULE) $(GO) list -e -compiled -test=true -export=false -deps=true -find=false -tags= -- ./... > /dev/null
	GO111MODULE=$(GO111MODULE) $(GOLANGCI_LINT) run $(GOLANGCI_LINT_OPTS) $(pkgs)
else
	$(GOLANGCI_LINT) run $(pkgs)
endif
endif

#ifdef GOLANGCI_LINT
#$(GOLANGCI_LINT):
#	mkdir -p $(FIRST_GOPATH)/bin
#	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/$(GOLANGCI_LINT_VERSION)/install.sh \
#		| sed -e '/install -d/d' \
#		| sh -s -- -b $(FIRST_GOPATH)/bin $(GOLANGCI_LINT_VERSION)
#endif
