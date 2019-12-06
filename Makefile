# Ensure GOBIN is not set during build so that latplot is installed to the correct path
unexport GOBIN
GO                 ?= go
GOFMT              ?= $(GO)fmt
FIRST_GOPATH       := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
LATPLOT_BIN        := ~/.local/bin/latplot
LATPLOT_VERSION    ?= 0.0.1
GOLANGCI_LINT      :=
GOLANGCI_LINT_OPTS ?=
ifeq ($(GOHOSTOS),$(filter $(GOHOSTOS),linux darwin))
ifeq ($(GOHOSTARCH),$(filter $(GOHOSTARCH),amd64 i386))
	GOLANGCI_LINT := $(shell which golangci-lint)
endif
endif


.PHONY: common-all
#common-all: precheck style check_license lint unused build test
common-all: common-lint test build

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

PREFIX             ?= $(shell pwd)

.PHONY: test
test:
	@echo ">> building latplot binaries"
	GO111MODULE=$(GO111MODULE) go test ./...

.PHONY: build
#build: assets common-build
build: latplot
	@echo ">> building latplot binaries"
	GO111MODULE=$(GO111MODULE) go build -o $(LATPLOT_BIN) cmd/latency-plot/main.go

.PHONY: latplot
latplot: $(LATPLOT)

$(LATPLOT):

#	$(eval LATPLOT_TMP := $(shell mktemp -d))
#	curl -s -L $(LATPLOT_URL) | tar -xvzf - -C $(PROMU_TMP)
#	mkdir -p $(FIRST_GOPATH)/bin
#	cp $(PROMU_TMP)/promu-$(PROMU_VERSION).$(GO_BUILD_PLATFORM)/promu $(FIRST_GOPATH)/bin/promu
#	rm -r $(PROMU_TMP)
