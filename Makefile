.PHONY: test test-property test-preview-performance test-mutation test-mutation-dry typst

GREMLINS ?= go run github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0

# Pinned Typst toolchain for the typst backend. `make typst` downloads the
# release tarball into ./bin, verifies its sha256 and extracts bin/typst.
# Overriding TYPST_VERSION requires the matching TYPST_SHA256.
TYPST_VERSION ?= 0.15.1
TYPST_SHA256 ?= a6d077d0a95eed5a2eba715b2dae06be954f624ccbf85758a03f389ded33118c
TYPST_TARGET ?= typst-x86_64-unknown-linux-musl
TYPST_URL ?= https://github.com/typst/typst/releases/download/v$(TYPST_VERSION)/$(TYPST_TARGET).tar.xz
TYPST_TARBALL := bin/$(TYPST_TARGET)-$(TYPST_VERSION).tar.xz

typst: bin/typst

bin/typst: $(TYPST_TARBALL)
	tar -xJf $(TYPST_TARBALL) -C bin --strip-components=1 $(TYPST_TARGET)/typst
	touch bin/typst
	bin/typst --version

$(TYPST_TARBALL):
	mkdir -p bin
	curl -fsSL -o $@.part $(TYPST_URL)
	echo "$(TYPST_SHA256)  $@.part" | sha256sum -c -
	mv $@.part $@

test:
	go test -count=1 ./...

test-property:
	go test -count=1 -run Property ./pkg/config ./pkg/api/domain/generation ./internal/autopdf/domain/options ./pkg/template/strict

test-preview-performance:
	go test -count=1 -run 'LatencyBudget' ./pkg/preview ./pkg/api/rest
	go test -run '^$$' -bench 'Compose(Cold|Warm)100|LaTeXPreview(Warm|Focused)|PreviewEventSerialization' -benchtime=3x ./pkg/composition ./pkg/preview ./pkg/api/rest

test-mutation:
	$(GREMLINS) unleash ./pkg/config
	$(GREMLINS) unleash ./pkg/api/domain/generation
	$(GREMLINS) unleash ./internal/autopdf/domain/options
	$(GREMLINS) unleash ./pkg/template/strict

test-mutation-dry:
	$(GREMLINS) unleash ./pkg/config --dry-run
	$(GREMLINS) unleash ./pkg/api/domain/generation --dry-run
	$(GREMLINS) unleash ./internal/autopdf/domain/options --dry-run
	$(GREMLINS) unleash ./pkg/template/strict --dry-run
