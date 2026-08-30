package beamer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/composition"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/document"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/latex"
)

var (
	ErrInvalidManifest = errors.New("invalid render manifest")
	ErrNestedSection   = errors.New("section fragments must be document roots")
)

// Package is one \usepackage line.
type Package struct {
	Name    string
	Options string
}

// DefaultPackages are the packages the bundled renderers require.
var DefaultPackages = []Package{
	{Name: "fontenc", Options: "T1"},
	{Name: "inputenc", Options: "utf8"},
	{Name: "graphicx"},
	{Name: "listings"},
	{Name: "tcolorbox"},
	{Name: "ulem", Options: "normalem"},
	{Name: "xcolor"},
	{Name: "hyperref"},
}

// DefaultPreamble is the style setup the bundled renderers assume.
// Options.Preamble replaces it.
const DefaultPreamble = `\tcbuselibrary{skins}
\lstset{basicstyle=\ttfamily\footnotesize,breaklines=true,columns=flexible,showstringspaces=false}
\setbeamertemplate{navigation symbols}{}
`

// Options controls the document shape, its styling, and focused rebuilds.
// Every field is inert data; Project performs no I/O.
type Options struct {
	// DocumentClass defaults to "beamer".
	DocumentClass string
	// ClassOptions defaults to AspectRatio alone.
	ClassOptions []string
	// Packages defaults to DefaultPackages. A non-nil empty slice loads none.
	Packages []Package
	// Preamble replaces DefaultPreamble when set.
	Preamble string
	// Theme is the Beamer theme name. Empty selects a plain one.
	Theme string
	// ColorTheme is the Beamer colour theme, or a style file's colour theme
	// generated from plato's design tokens.
	ColorTheme string
	// StyleFile is a .sty to \usepackage after the theme, which is how a
	// token-generated palette reaches the document.
	StyleFile string
	// TrustedPreamble is appended verbatim. It is trusted: callers must not
	// route document content through it.
	TrustedPreamble []byte
	// FocusSections limits the build to these section components via
	// \includeonly. Each must name a root section.
	FocusSections []string
	// GraphicsPath is prepended to image lookups, letting a projection
	// reference assets that live outside the compile workspace.
	GraphicsPath string
	// Title, Author and Date populate \titlepage when Title is set.
	Title, Author, Date string
	// ShowNotes renders speaker notes on a second screen.
	ShowNotes bool
	// AspectRatio is the Beamer aspectratio option; 169 by default.
	AspectRatio string
}

// Project assembles a manifest into a compilable Beamer document, returning
// the same latex.Projection the LaTeX target produces.
//
// Each root section becomes one \include'd file; flow and artifact fragments
// are spliced inline into the frame that contains them. Roots must all be
// section-mode fragments. Project is pure.
func Project(manifest composition.RenderManifest, options Options) (latex.Projection, error) {
	fragments, err := indexManifest(manifest)
	if err != nil {
		return latex.Projection{}, err
	}
	if err := validateTree(manifest.RootOrder, fragments); err != nil {
		return latex.Projection{}, err
	}

	projection := latex.Projection{
		Main:      "main.tex",
		SourceMap: make(map[string]latex.SourceLocation, len(manifest.RootOrder)),
	}
	paths := make(map[string]string, len(manifest.RootOrder))
	for _, id := range manifest.RootOrder {
		paths[id] = "frames/" + safeName(id) + "-" + shortHash(fragments[id].Hash)
	}

	for _, id := range manifest.RootOrder {
		content, err := assemble(id, fragments)
		if err != nil {
			return latex.Projection{}, err
		}
		projection.Files = append(projection.Files, latex.File{Path: paths[id] + ".tex", Content: content})
		projection.SourceMap[id] = latex.SourceLocation{Path: paths[id] + ".tex", Line: 1}
	}

	main, err := mainSource(manifest, paths, options)
	if err != nil {
		return latex.Projection{}, err
	}
	projection.Files = append(projection.Files, latex.File{Path: projection.Main, Content: main})
	sort.Slice(projection.Files, func(i, j int) bool { return projection.Files[i].Path < projection.Files[j].Path })
	projection.Hash = hashProjection(projection)
	return projection, nil
}

// Projector adapts Project to the engine's projector seam. Options are fixed
// at construction; a request contributes only its focus sections.
type Projector struct {
	options Options
}

// NewProjector returns a projector that renders every manifest with options.
func NewProjector(options Options) *Projector { return &Projector{options: options} }

// ProjectManifest satisfies the engine's projector seam.
func (projector *Projector) ProjectManifest(manifest composition.RenderManifest, focusSections []string) (latex.Projection, error) {
	options := projector.options
	options.FocusSections = append([]string(nil), focusSections...)
	return Project(manifest, options)
}

func indexManifest(manifest composition.RenderManifest) (map[string]composition.Fragment, error) {
	fragments := make(map[string]composition.Fragment, len(manifest.Fragments))
	for _, fragment := range manifest.Fragments {
		if strings.TrimSpace(fragment.ComponentID) == "" || !fragment.Mode.Valid() {
			return nil, fmt.Errorf("%w: fragment identity or mode", ErrInvalidManifest)
		}
		if _, exists := fragments[fragment.ComponentID]; exists {
			return nil, fmt.Errorf("%w: duplicate fragment %q", ErrInvalidManifest, fragment.ComponentID)
		}
		fragments[fragment.ComponentID] = fragment
	}
	return fragments, nil
}

func validateTree(roots []string, fragments map[string]composition.Fragment) error {
	rootSet := make(map[string]struct{}, len(roots))
	for _, root := range roots {
		if _, exists := rootSet[root]; exists {
			return fmt.Errorf("%w: duplicate root %q", ErrInvalidManifest, root)
		}
		rootSet[root] = struct{}{}
		fragment, exists := fragments[root]
		if !exists {
			return fmt.Errorf("%w: missing root %q", ErrInvalidManifest, root)
		}
		if fragment.Mode != document.Section {
			return fmt.Errorf("%w: root %q is not a section; a Beamer document is a sequence of frames", ErrInvalidManifest, root)
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

// assemble splices a fragment's rendered children into it at the first
// ChildrenMarker, appending them when the fragment carries no marker.
func assemble(id string, fragments map[string]composition.Fragment) ([]byte, error) {
	fragment, exists := fragments[id]
	if !exists {
		return nil, fmt.Errorf("%w: missing child %q", ErrInvalidManifest, id)
	}
	var children bytes.Buffer
	for _, child := range fragment.Children {
		content, err := assemble(child, fragments)
		if err != nil {
			return nil, err
		}
		children.Write(content)
	}
	content := fragment.Content
	if bytes.Contains(content, []byte(ChildrenMarker)) {
		return bytes.Replace(content, []byte(ChildrenMarker), children.Bytes(), 1), nil
	}
	return append(append([]byte(nil), content...), children.Bytes()...), nil
}

func mainSource(manifest composition.RenderManifest, paths map[string]string, options Options) ([]byte, error) {
	class := safeOption(options.DocumentClass)
	if class == "" {
		class = "beamer"
	}
	classOptions := options.ClassOptions
	if classOptions == nil {
		aspect := options.AspectRatio
		if strings.TrimSpace(aspect) == "" {
			aspect = "169"
		}
		classOptions = []string{"aspectratio=" + aspect}
	}
	safeClassOptions := make([]string, 0, len(classOptions))
	for _, option := range classOptions {
		if safe := safeArgument(option); safe != "" {
			safeClassOptions = append(safeClassOptions, safe)
		}
	}

	packages := options.Packages
	if packages == nil {
		packages = DefaultPackages
	}
	preamble := options.Preamble
	if preamble == "" {
		preamble = DefaultPreamble
	}

	var source bytes.Buffer
	source.WriteString("\\documentclass")
	if len(safeClassOptions) > 0 {
		source.WriteString("[" + strings.Join(safeClassOptions, ",") + "]")
	}
	source.WriteString("{" + class + "}\n")
	for _, item := range packages {
		name := safeOption(item.Name)
		if name == "" {
			continue
		}
		source.WriteString("\\usepackage")
		if item.Options != "" {
			source.WriteString("[" + safeArgument(item.Options) + "]")
		}
		source.WriteString("{" + name + "}\n")
	}
	source.WriteString(preamble)
	if !strings.HasSuffix(preamble, "\n") {
		source.WriteByte('\n')
	}
	if theme := safeOption(options.Theme); theme != "" {
		source.WriteString("\\usetheme{" + theme + "}\n")
	}
	if colorTheme := safeOption(options.ColorTheme); colorTheme != "" {
		source.WriteString("\\usecolortheme{" + colorTheme + "}\n")
	}
	if style := safeOption(strings.TrimSuffix(options.StyleFile, ".sty")); style != "" {
		source.WriteString("\\usepackage{" + style + "}\n")
	}
	if options.GraphicsPath != "" {
		path := strings.TrimSuffix(escapePath(options.GraphicsPath), "/") + "/"
		// Both are required: \graphicspath serves \includegraphics,
		// \input@path serves the \IfFileExists guard around it.
		source.WriteString("\\graphicspath{{" + path + "}}\n")
		source.WriteString("\\makeatletter\\def\\input@path{{" + path + "}}\\makeatother\n")
	}
	if options.ShowNotes {
		source.WriteString("\\setbeameroption{show notes}\n")
	}
	if options.Title != "" {
		source.WriteString("\\title{" + escape(options.Title) + "}\n")
	}
	if options.Author != "" {
		source.WriteString("\\author{" + escape(options.Author) + "}\n")
	}
	if options.Date != "" {
		source.WriteString("\\date{" + escape(options.Date) + "}\n")
	}
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
			path, exists := paths[id]
			if !exists {
				return nil, fmt.Errorf("%w: focus %q is not a section", ErrInvalidManifest, id)
			}
			if _, duplicate := seen[id]; duplicate {
				continue
			}
			seen[id] = struct{}{}
			focus = append(focus, path)
		}
		sort.Strings(focus)
		source.WriteString("\\includeonly{" + strings.Join(focus, ",") + "}\n")
	}

	source.WriteString("\\begin{document}\n")
	if options.Title != "" {
		source.WriteString("\\frame{\\titlepage}\n")
	}
	for _, id := range manifest.RootOrder {
		source.WriteString("\\include{" + paths[id] + "}\n")
	}
	source.WriteString("\\end{document}\n")
	return source.Bytes(), nil
}

// safeOption reduces a preamble word — a class, package or theme name — to
// letters, digits, '-' and '.'.
func safeOption(value string) string {
	return keep(value, func(character rune) bool {
		return unicode.IsLetter(character) || unicode.IsDigit(character) ||
			character == '-' || character == '.'
	})
}

// safeArgument reduces an option list to safeOption's characters plus the
// separators '=', ',', ':' and space.
func safeArgument(value string) string {
	return keep(value, func(character rune) bool {
		return unicode.IsLetter(character) || unicode.IsDigit(character) ||
			strings.ContainsRune("-.=,: ", character)
	})
}

func keep(value string, allowed func(rune) bool) string {
	var result strings.Builder
	for _, character := range strings.TrimSpace(value) {
		if allowed(character) {
			result.WriteRune(character)
		}
	}
	return result.String()
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
		return "frame"
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
	digest := sha256.New()
	digest.Write([]byte(projection.Main))
	for _, file := range projection.Files {
		digest.Write([]byte(file.Path))
		digest.Write([]byte{0})
		digest.Write(file.Content)
		digest.Write([]byte{0})
	}
	return hex.EncodeToString(digest.Sum(nil))
}
