.PHONY: test test-property test-preview-performance test-mutation test-mutation-dry

GREMLINS ?= go run github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0

test:
	go test -count=1 ./...

test-property:
	go test -count=1 -run Property ./pkg/config ./pkg/api/domain/generation ./internal/autopdf/domain/options

test-preview-performance:
	go test -count=1 -run 'LatencyBudget' ./pkg/preview ./pkg/api/rest
	go test -run '^$$' -bench 'Compose(Cold|Warm)100|LaTeXPreview(Warm|Focused)|PreviewEventSerialization' -benchtime=3x ./pkg/composition ./pkg/preview ./pkg/api/rest

test-mutation:
	$(GREMLINS) unleash ./pkg/config
	$(GREMLINS) unleash ./pkg/api/domain/generation
	$(GREMLINS) unleash ./internal/autopdf/domain/options

test-mutation-dry:
	$(GREMLINS) unleash ./pkg/config --dry-run
	$(GREMLINS) unleash ./pkg/api/domain/generation --dry-run
	$(GREMLINS) unleash ./internal/autopdf/domain/options --dry-run
