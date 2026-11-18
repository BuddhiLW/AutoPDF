// Package typst projects composed render manifests into a compilable Typst
// workspace: a main .typ file, one .typ file per fragment, and inert asset
// bindings an effectful collector materializes.
package typst

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/component"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/composition"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/document"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/latex"
)

var (
	ErrInvalidManifest      = errors.New("invalid render manifest")
	ErrNestedSection        = errors.New("section fragments must be document roots")
	ErrUnsupportedAsset     = errors.New("unsupported artifact media type")
	ErrUnsupportedComponent = errors.New("component has no typst projection")
)

// UnsupportedComponentError names the component the Typst projector refused
// and the dispatch key it carried. It matches ErrUnsupportedComponent.
type UnsupportedComponentError struct {
	ComponentID string
	Key         component.Key
}

func (failure *UnsupportedComponentError) Error() string {
	return fmt.Sprintf("%s: component %q (%s)", ErrUnsupportedComponent, failure.ComponentID, failure.Key)
}

func (failure *UnsupportedComponentError) Is(target error) bool {
	return target == ErrUnsupportedComponent
}

// DiagnosticPattern matches `typst compile --diagnostic-format short` lines:
// submatch 1 is the file (root-relative, "/fragments/x.typ"), 2 the line and
// 3 the severity and message.
var DiagnosticPattern = regexp.MustCompile(`(?m)([^:\n]+\.typ):(\d+):\d+:\s*([^\n]+)`)

// MainFile is the workspace-relative entry point of every projection.
const MainFile = "main.typ"

// Options controls trusted document-level Typst and focused section builds.
type Options struct {
	// TrustedPreamble is emitted verbatim at the top of main.typ (set and
	// show rules). Callers must not route document content through it.
	TrustedPreamble []byte
	// FocusSections limits the build to these root sections; flow and
	// artifact roots are always kept, as LaTeX's \includeonly does.
	FocusSections []string
	// Supported lists the dispatch keys whose fragments hold Typst markup.
	// Nil selects DefaultKeys. Any other fragment is refused by name.
	Supported []component.Key
}

// Project maps composition modes to Typst source boundaries:
// flow -> #include, root section -> page-isolated #include, artifact ->
// #image of a bound asset. The result shares latex.Projection's shape so the
// projection generator and preview pipeline carry it unchanged.
func Project(manifest composition.RenderManifest, options Options) (latex.Projection, error) {
	fragments, paths, err := indexManifest(manifest, options.Supported)
	if err != nil {
		return latex.Projection{}, err
	}
	if err := validateTree(manifest.RootOrder, fragments); err != nil {
		return latex.Projection{}, err
	}

	projection := latex.Projection{
		Main:      MainFile,
		SourceMap: make(map[string]latex.SourceLocation, len(fragments)),
	}
	ids := make([]string, 0, len(fragments))
	for id := range fragments {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		content, bindings, err := fragmentSource(fragments[id], paths)
		if err != nil {
			return latex.Projection{}, err
		}
		projection.Files = append(projection.Files, latex.File{Path: paths[id], Content: content})
		projection.Assets = append(projection.Assets, bindings...)
		projection.SourceMap[id] = latex.SourceLocation{Path: paths[id], Line: 1}
	}

	main, err := mainSource(manifest.RootOrder, fragments, paths, options)
	if err != nil {
		return latex.Projection{}, err
	}
	projection.Files = append(projection.Files, latex.File{Path: projection.Main, Content: main})
	sort.Slice(projection.Files, func(i, j int) bool { return projection.Files[i].Path < projection.Files[j].Path })
	sort.Slice(projection.Assets, func(i, j int) bool { return projection.Assets[i].Path < projection.Assets[j].Path })
	projection.Hash = hashProjection(projection)
	return projection, nil
}

// Projector adapts Project to the document engine's projector seam.
type Projector struct {
	options Options
}

// NewProjector returns a projector that renders every manifest with options;
// a request contributes only its focus sections.
func NewProjector(options Options) *Projector {
	options.TrustedPreamble = append([]byte(nil), options.TrustedPreamble...)
	options.Supported = append([]component.Key(nil), options.Supported...)
	if len(options.Supported) == 0 {
		options.Supported = nil
	}
	return &Projector{options: options}
}

// ProjectManifest satisfies the engine's projector seam.
func (projector *Projector) ProjectManifest(manifest composition.RenderManifest, focusSections []string) (latex.Projection, error) {
	options := projector.options
	options.FocusSections = append([]string(nil), focusSections...)
	return Project(manifest, options)
}

func indexManifest(manifest composition.RenderManifest, supported []component.Key) (map[string]composition.Fragment, map[string]string, error) {
	if supported == nil {
		supported = DefaultKeys()
	}
	accepted := make(map[component.Key]struct{}, len(supported))
	for _, key := range supported {
		accepted[key] = struct{}{}
	}
	fragments := make(map[string]composition.Fragment, len(manifest.Fragments))
	paths := make(map[string]string, len(manifest.Fragments))
	for _, fragment := range manifest.Fragments {
		if strings.TrimSpace(fragment.ComponentID) == "" || !fragment.Mode.Valid() {
			return nil, nil, fmt.Errorf("%w: fragment identity or mode", ErrInvalidManifest)
		}
		if _, exists := fragments[fragment.ComponentID]; exists {
			return nil, nil, fmt.Errorf("%w: duplicate fragment %q", ErrInvalidManifest, fragment.ComponentID)
		}
		if _, ok := accepted[fragment.Key]; !ok {
			return nil, nil, &UnsupportedComponentError{ComponentID: fragment.ComponentID, Key: fragment.Key}
		}
		fragments[fragment.ComponentID] = fragment
		paths[fragment.ComponentID] = "fragments/" + safeName(fragment.ComponentID) + "-" + shortHash(fragment.Hash) + ".typ"
	}
	return fragments, paths, nil
}

func validateTree(roots []string, fragments map[string]composition.Fragment) error {
	rootSet := make(map[string]struct{}, len(roots))
	for _, root := range roots {
		if _, exists := rootSet[root]; exists {
			return fmt.Errorf("%w: duplicate root %q", ErrInvalidManifest, root)
		}
		rootSet[root] = struct{}{}
		if _, exists := fragments[root]; !exists {
			return fmt.Errorf("%w: missing root %q", ErrInvalidManifest, root)
		}
	}
	state := make(map[string]uint8, len(fragments))
	var visit func(string, bool) error
	visit = func(id string, isRoot bool) error {
		fragment, exists := fragments[id]
		if !exists {
			return fmt.Errorf("%w: missing child %q", ErrInvalidManifest, id)
		}
		if fragment.Mode == document.Section && !isRoot {
			return fmt.Errorf("%w: %s", ErrNestedSection, id)
		}
		if state[id] == 1 {
			return fmt.Errorf("%w: component cycle at %q", ErrInvalidManifest, id)
		}
		if state[id] == 2 {
			return nil
		}
		state[id] = 1
		for _, child := range fragment.Children {
			if err := visit(child, false); err != nil {
				return err
			}
		}
		state[id] = 2
		return nil
	}
	for _, root := range roots {
		if err := visit(root, true); err != nil {
			return err
		}
	}
	if len(state) != len(fragments) {
		return fmt.Errorf("%w: orphan fragments", ErrInvalidManifest)
	}
	return nil
}

func include(relative string) string {
	return "#include \"/" + relative + "\"\n\n"
}

func fragmentSource(fragment composition.Fragment, paths map[string]string) ([]byte, []latex.AssetBinding, error) {
	var source bytes.Buffer
	source.Write(fragment.Content)
	if len(fragment.Content) > 0 && fragment.Content[len(fragment.Content)-1] != '\n' {
		source.WriteByte('\n')
	}
	bindings := make([]latex.AssetBinding, 0, len(fragment.Assets))
	if fragment.Mode == document.Artifact {
		for _, asset := range fragment.Assets {
			binding, directive, err := assetDirective(fragment.ComponentID, asset)
			if err != nil {
				return nil, nil, err
			}
			bindings = append(bindings, binding)
			source.WriteString(directive)
			source.WriteByte('\n')
		}
	}
	for _, child := range fragment.Children {
		childPath, exists := paths[child]
		if !exists {
			return nil, nil, fmt.Errorf("%w: missing child %q", ErrInvalidManifest, child)
		}
		source.WriteString(include(childPath))
	}
	return source.Bytes(), bindings, nil
}

func mainSource(roots []string, fragments map[string]composition.Fragment, paths map[string]string, options Options) ([]byte, error) {
	var source bytes.Buffer
	if len(options.TrustedPreamble) > 0 {
		source.Write(options.TrustedPreamble)
		if options.TrustedPreamble[len(options.TrustedPreamble)-1] != '\n' {
			source.WriteByte('\n')
		}
	}
	var focus map[string]struct{}
	if len(options.FocusSections) > 0 {
		focus = make(map[string]struct{}, len(options.FocusSections))
		for _, id := range options.FocusSections {
			fragment, exists := fragments[id]
			if !exists || fragment.Mode != document.Section {
				return nil, fmt.Errorf("%w: focus %q is not a section", ErrInvalidManifest, id)
			}
			focus[id] = struct{}{}
		}
	}
	for _, id := range roots {
		if fragments[id].Mode != document.Section {
			source.WriteString(include(paths[id]))
			continue
		}
		if focus != nil {
			if _, kept := focus[id]; !kept {
				continue
			}
		}
		source.WriteString("#pagebreak(weak: true)\n")
		source.WriteString(include(paths[id]))
		source.WriteString("#pagebreak(weak: true)\n")
	}
	return source.Bytes(), nil
}

func assetDirective(componentID string, asset document.AssetRef) (latex.AssetBinding, string, error) {
	extension, err := assetFormat(asset)
	if err != nil {
		return latex.AssetBinding{}, "", fmt.Errorf("component %s asset %s: %w", componentID, asset.ID, err)
	}
	digest := sha256.Sum256([]byte(componentID + "\x00" + asset.ID + "\x00" + asset.URI + "\x00" + asset.MediaType + "\x00" + asset.Digest))
	assetPath := "assets/" + safeName(componentID) + "-" + safeName(asset.ID) + "-" + hex.EncodeToString(digest[:4]) + extension
	binding := latex.AssetBinding{ComponentID: componentID, AssetID: asset.ID, SourceURI: asset.URI, MediaType: asset.MediaType, Digest: asset.Digest, Path: assetPath}
	return binding, "#image(\"/" + assetPath + "\")", nil
}

// imageFormats are the media types Typst's image element decodes. PDF is
// absent on purpose: Typst embeds one PDF page per image, so a multi-page
// artifact would be truncated silently.
var imageFormats = map[string]string{
	"image/png":     ".png",
	"image/jpeg":    ".jpg",
	"image/gif":     ".gif",
	"image/svg+xml": ".svg",
	"image/webp":    ".webp",
}

var extensionMediaTypes = map[string]string{
	".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg",
	".gif": "image/gif", ".svg": "image/svg+xml", ".webp": "image/webp",
	".pdf": "application/pdf",
}

func assetFormat(asset document.AssetRef) (string, error) {
	mediaType := strings.ToLower(strings.TrimSpace(strings.Split(asset.MediaType, ";")[0]))
	if mediaType == "" {
		mediaType = extensionMediaTypes[strings.ToLower(path.Ext(asset.URI))]
	}
	extension, ok := imageFormats[mediaType]
	if !ok {
		return "", fmt.Errorf("%w %q", ErrUnsupportedAsset, asset.MediaType)
	}
	return extension, nil
}

func safeName(value string) string {
	var result strings.Builder
	lastDash := false
	for _, character := range strings.ToLower(value) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			result.WriteRune(character)
			lastDash = false
		} else if !lastDash && result.Len() > 0 {
			result.WriteByte('-')
			lastDash = true
		}
	}
	name := strings.Trim(result.String(), "-")
	if name == "" {
		return "component"
	}
	return name
}

func shortHash(value string) string {
	if len(value) >= 12 {
		return value[:12]
	}
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:6])
}

func hashProjection(projection latex.Projection) string {
	hash := sha256.New()
	_, _ = hash.Write([]byte(projection.Main))
	_, _ = hash.Write([]byte{0})
	for _, file := range projection.Files {
		_, _ = hash.Write([]byte(file.Path))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write(file.Content)
		_, _ = hash.Write([]byte{0})
	}
	for _, asset := range projection.Assets {
		_, _ = hash.Write([]byte(asset.ComponentID + "\x00" + asset.AssetID + "\x00" + asset.SourceURI + "\x00" + asset.MediaType + "\x00" + asset.Digest + "\x00" + asset.Path))
	}
	return hex.EncodeToString(hash.Sum(nil))
}
