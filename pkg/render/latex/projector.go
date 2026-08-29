package latex

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
	"unicode"

	"github.com/BuddhiLW/AutoPDF/pkg/composition"
	"github.com/BuddhiLW/AutoPDF/pkg/document"
)

var (
	ErrInvalidManifest  = errors.New("invalid render manifest")
	ErrNestedSection    = errors.New("section fragments must be document roots")
	ErrUnsupportedAsset = errors.New("unsupported artifact media type")
)

// Options controls trusted document-level LaTeX and focused section builds.
type Options struct {
	TrustedPreamble []byte
	FocusSections   []string
}

// File is one generated, workspace-relative source file.
type File struct {
	Path    string `json:"path"`
	Content []byte `json:"content"`
}

// AssetBinding tells an effectful collector where an inert asset must land.
type AssetBinding struct {
	ComponentID string `json:"componentId"`
	AssetID     string `json:"assetId"`
	SourceURI   string `json:"sourceUri"`
	MediaType   string `json:"mediaType,omitempty"`
	Digest      string `json:"digest,omitempty"`
	Path        string `json:"path"`
}

// SourceLocation maps editor identity to generated LaTeX.
type SourceLocation struct {
	Path string `json:"path"`
	Line int    `json:"line"`
}

// Projection is a complete, deterministic input to an effectful compiler.
type Projection struct {
	Main      string                    `json:"main"`
	Files     []File                    `json:"files"`
	Assets    []AssetBinding            `json:"assets"`
	SourceMap map[string]SourceLocation `json:"sourceMap"`
	Hash      string                    `json:"hash"`
}

// Project maps domain composition modes to safe LaTeX source boundaries:
// flow -> input, top-level section -> include, artifact -> embedded asset.
func Project(manifest composition.RenderManifest, options Options) (Projection, error) {
	fragments, paths, err := indexManifest(manifest)
	if err != nil {
		return Projection{}, err
	}
	if err := validateTree(manifest.RootOrder, fragments); err != nil {
		return Projection{}, err
	}

	projection := Projection{
		Main:      "main.tex",
		SourceMap: make(map[string]SourceLocation, len(fragments)),
	}
	ids := make([]string, 0, len(fragments))
	for id := range fragments {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		fragment := fragments[id]
		content, bindings, err := fragmentSource(fragment, paths)
		if err != nil {
			return Projection{}, err
		}
		projection.Files = append(projection.Files, File{Path: paths[id], Content: content})
		projection.Assets = append(projection.Assets, bindings...)
		projection.SourceMap[id] = SourceLocation{Path: paths[id], Line: 1}
	}

	mainSource, err := mainSource(manifest.RootOrder, fragments, paths, options)
	if err != nil {
		return Projection{}, err
	}
	projection.Files = append(projection.Files, File{Path: projection.Main, Content: mainSource})
	sort.Slice(projection.Files, func(i, j int) bool { return projection.Files[i].Path < projection.Files[j].Path })
	sort.Slice(projection.Assets, func(i, j int) bool { return projection.Assets[i].Path < projection.Assets[j].Path })
	projection.Hash = hashProjection(projection)
	return projection, nil
}

func indexManifest(manifest composition.RenderManifest) (map[string]composition.Fragment, map[string]string, error) {
	fragments := make(map[string]composition.Fragment, len(manifest.Fragments))
	paths := make(map[string]string, len(manifest.Fragments))
	for _, fragment := range manifest.Fragments {
		if strings.TrimSpace(fragment.ComponentID) == "" || !fragment.Mode.Valid() {
			return nil, nil, fmt.Errorf("%w: fragment identity or mode", ErrInvalidManifest)
		}
		if _, exists := fragments[fragment.ComponentID]; exists {
			return nil, nil, fmt.Errorf("%w: duplicate fragment %q", ErrInvalidManifest, fragment.ComponentID)
		}
		fragments[fragment.ComponentID] = fragment
		paths[fragment.ComponentID] = "fragments/" + safeName(fragment.ComponentID) + "-" + shortHash(fragment.Hash) + ".tex"
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

func fragmentSource(fragment composition.Fragment, paths map[string]string) ([]byte, []AssetBinding, error) {
	var source bytes.Buffer
	source.Write(fragment.Content)
	if len(fragment.Content) > 0 && fragment.Content[len(fragment.Content)-1] != '\n' {
		source.WriteByte('\n')
	}
	bindings := make([]AssetBinding, 0, len(fragment.Assets))
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
		source.WriteString("\\input{")
		source.WriteString(strings.TrimSuffix(childPath, ".tex"))
		source.WriteString("}\n")
	}
	return source.Bytes(), bindings, nil
}

func mainSource(roots []string, fragments map[string]composition.Fragment, paths map[string]string, options Options) ([]byte, error) {
	var source bytes.Buffer
	source.WriteString("\\documentclass{article}\n\\usepackage{graphicx}\n\\usepackage{pdfpages}\n")
	if len(options.TrustedPreamble) > 0 {
		source.Write(options.TrustedPreamble)
		if options.TrustedPreamble[len(options.TrustedPreamble)-1] != '\n' {
			source.WriteByte('\n')
		}
	}
	if len(options.FocusSections) > 0 {
		focus := make([]string, 0, len(options.FocusSections))
		seen := make(map[string]struct{}, len(options.FocusSections))
		for _, id := range options.FocusSections {
			fragment, exists := fragments[id]
			if !exists || fragment.Mode != document.Section {
				return nil, fmt.Errorf("%w: focus %q is not a section", ErrInvalidManifest, id)
			}
			if _, exists := seen[id]; exists {
				continue
			}
			seen[id] = struct{}{}
			focus = append(focus, strings.TrimSuffix(paths[id], ".tex"))
		}
		sort.Strings(focus)
		source.WriteString("\\includeonly{")
		source.WriteString(strings.Join(focus, ","))
		source.WriteString("}\n")
	}
	source.WriteString("\\begin{document}\n")
	for _, id := range roots {
		fragment := fragments[id]
		command := "input"
		if fragment.Mode == document.Section {
			command = "include"
		}
		source.WriteByte('\\')
		source.WriteString(command)
		source.WriteByte('{')
		source.WriteString(strings.TrimSuffix(paths[id], ".tex"))
		source.WriteString("}\n")
	}
	source.WriteString("\\end{document}\n")
	return source.Bytes(), nil
}

func assetDirective(componentID string, asset document.AssetRef) (AssetBinding, string, error) {
	extension, command, err := assetFormat(asset)
	if err != nil {
		return AssetBinding{}, "", fmt.Errorf("component %s asset %s: %w", componentID, asset.ID, err)
	}
	digest := sha256.Sum256([]byte(componentID + "\x00" + asset.ID + "\x00" + asset.URI + "\x00" + asset.MediaType + "\x00" + asset.Digest))
	assetPath := "assets/" + safeName(componentID) + "-" + safeName(asset.ID) + "-" + hex.EncodeToString(digest[:4]) + extension
	binding := AssetBinding{ComponentID: componentID, AssetID: asset.ID, SourceURI: asset.URI, MediaType: asset.MediaType, Digest: asset.Digest, Path: assetPath}
	return binding, command + "{" + assetPath + "}", nil
}

func assetFormat(asset document.AssetRef) (string, string, error) {
	mediaType := strings.ToLower(strings.TrimSpace(strings.Split(asset.MediaType, ";")[0]))
	if mediaType == "" {
		switch strings.ToLower(path.Ext(asset.URI)) {
		case ".pdf":
			mediaType = "application/pdf"
		case ".png":
			mediaType = "image/png"
		case ".jpg", ".jpeg":
			mediaType = "image/jpeg"
		}
	}
	switch mediaType {
	case "application/pdf":
		return ".pdf", "\\includepdf[pages=-]", nil
	case "image/png":
		return ".png", "\\includegraphics", nil
	case "image/jpeg":
		return ".jpg", "\\includegraphics", nil
	default:
		return "", "", fmt.Errorf("%w %q", ErrUnsupportedAsset, asset.MediaType)
	}
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

func hashProjection(projection Projection) string {
	hash := sha256.New()
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
