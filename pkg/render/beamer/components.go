// Package beamer registers a `beamer` rendering target for the component
// catalog and projects its manifests into a compilable Beamer document.
//
// The kinds implemented here are the vocabulary specified in
// docs/plato-integration.md.
package beamer

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/component"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/document"
)

// Target is the dispatch target these definitions register under.
const Target = "beamer"

// These tables are the extension points of the beamer target: each is an open
// set a caller may add to before building a catalog.
var (
	// MarkCommands name the LaTeX command wrapping a span that carries the
	// mark. Marks apply in sorted order.
	MarkCommands = map[string]string{
		"strong":    "textbf",
		"emph":      "emph",
		"code":      "texttt",
		"strike":    "sout",
		"highlight": "underline",
	}

	// CalloutTones map a callout's tone to its tcolorbox colours.
	CalloutTones = map[string]struct{ Background, Frame string }{
		"info": {"blue!5!white", "blue!60!black"},
		"warn": {"orange!10!white", "orange!70!black"},
		"ok":   {"green!8!white", "green!50!black"},
	}

	// DefaultTone is used when a callout names a tone this table lacks.
	DefaultTone = "info"

	// ListingLanguages map a lower-cased source language to a listings
	// package language. An absent key renders unhighlighted.
	ListingLanguages = map[string]string{
		"clojure": "Lisp", "clj": "Lisp", "cljs": "Lisp", "cljc": "Lisp",
		"lisp": "Lisp", "scheme": "Lisp", "elisp": "Lisp",
		"go": "Go", "golang": "Go",
		"python": "Python", "py": "Python",
		"javascript": "Java", "js": "Java", "typescript": "Java", "ts": "Java",
		"java": "Java", "c": "C", "c++": "C++", "cpp": "C++",
		"haskell": "Haskell", "hs": "Haskell",
		"ruby": "Ruby", "rb": "Ruby",
		"rust": "C", "sql": "SQL",
		"bash": "bash", "sh": "bash", "shell": "bash", "zsh": "bash",
		"tex": "TeX", "latex": "TeX",
		"xml": "XML", "html": "XML",
	}

	// PrintableExtensions are the lower-cased file extensions that may reach
	// \includegraphics. Anything else renders as a placeholder.
	PrintableExtensions = map[string]bool{
		".pdf": true, ".png": true, ".jpg": true, ".jpeg": true, ".eps": true,
	}

	// UnprintableReasons label an unprintable extension on its placeholder.
	UnprintableReasons = map[string]string{
		".gif": "animated image",
		".svg": "vector image — convert to PDF at the asset boundary",
	}
)

// ChildrenMarker is where the projector splices a container's rendered
// children. A container renderer emits its opening and closing LaTeX around
// this marker; a fragment without it receives its children appended.
const ChildrenMarker = "%%autopdf:children%%"

// Definitions returns every component definition for the beamer target.
func Definitions() ([]component.Definition, error) {
	specs := []struct {
		kind     string
		mode     document.CompositionMode
		defaults document.Props
		validate component.Validator
		render   component.Renderer
	}{
		{kind: "section", mode: document.Section, render: renderSection},
		{kind: "text", mode: document.Flow, render: renderText},
		{kind: "span", mode: document.Flow, render: renderSpan},
		{kind: "link", mode: document.Flow, render: renderLink},
		{kind: "heading", mode: document.Flow, render: renderHeading},
		{kind: "bullets", mode: document.Flow, render: renderBullets, validate: validateBullets},
		{kind: "code", mode: document.Flow, render: renderCode},
		{kind: "quote", mode: document.Flow, render: renderQuote},
		{kind: "table", mode: document.Flow, render: renderTable, validate: validateTable},
		{kind: "image", mode: document.Flow, render: renderImage},
		{kind: "columns", mode: document.Flow, render: renderColumns},
		{kind: "cards", mode: document.Flow, render: renderCards},
		{kind: "card", mode: document.Flow, render: renderCard},
		{kind: "callout", mode: document.Flow, render: renderCallout},
		{kind: "kicker", mode: document.Flow, render: renderKicker},
		{kind: "media-placeholder", mode: document.Flow, render: renderMedia},
		{kind: "notes", mode: document.Flow, render: renderNotes},
		{kind: "rule", mode: document.Flow, render: renderRule},
		{kind: "scene", mode: document.Artifact, render: renderScene},
	}

	definitions := make([]component.Definition, 0, len(specs))
	for _, spec := range specs {
		validate := spec.validate
		if validate == nil {
			validate = noProblems
		}
		definition, err := component.NewDefinition(
			component.Key{Kind: spec.kind, Variant: "default", Target: Target},
			spec.mode,
			json.RawMessage(`{"type":"object"}`),
			spec.defaults,
			validate,
			spec.render,
		)
		if err != nil {
			return nil, fmt.Errorf("define beamer component %q: %w", spec.kind, err)
		}
		definitions = append(definitions, definition)
	}
	return definitions, nil
}

// Register installs every beamer definition into a catalog builder. It fails
// if any key is already registered; to override one kind, register a filtered
// subset of Definitions() together with your own.
func Register(builder *component.Builder) error {
	definitions, err := Definitions()
	if err != nil {
		return err
	}
	for _, definition := range definitions {
		if err := builder.Register(definition); err != nil {
			return err
		}
	}
	return nil
}

// Catalog freezes a catalog holding exactly the beamer definitions.
//
//	catalog, _ := beamer.Catalog()
//	engine, _ := api.NewDocumentEngine(catalog, api.DocumentEngineConfig{
//	    Target:    beamer.Target,
//	    Projector: beamer.NewProjector(beamer.Options{Theme: "Madrid"}),
//	})
func Catalog() (*component.Catalog, error) {
	builder := component.NewBuilder()
	if err := Register(builder); err != nil {
		return nil, err
	}
	return builder.Freeze()
}

func noProblems(document.Component) document.Problems { return nil }

// ── prop readers ────────────────────────────────────────────────────────────
//
// Props arrive as decoded JSON: numbers are json.Number, arrays are []any.
// Every reader returns the zero value for an absent or ill-typed prop and
// never panics.

func propString(node document.Component, key string) string {
	value, exists := node.Props[key]
	if !exists {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case bool:
		return fmt.Sprintf("%t", typed)
	default:
		return ""
	}
}

func propBool(node document.Component, key string) bool {
	value, _ := node.Props[key].(bool)
	return value
}

func propInt(node document.Component, key string, fallback int) int {
	value, exists := node.Props[key]
	if !exists {
		return fallback
	}
	number, ok := value.(json.Number)
	if !ok {
		return fallback
	}
	parsed, err := number.Int64()
	if err != nil {
		return fallback
	}
	return int(parsed)
}

func propStrings(node document.Component, key string) []string {
	values, ok := node.Props[key].([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		switch typed := value.(type) {
		case string:
			result = append(result, typed)
		case json.Number:
			result = append(result, typed.String())
		}
	}
	return result
}

func propRows(node document.Component, key string) [][]string {
	values, ok := node.Props[key].([]any)
	if !ok {
		return nil
	}
	rows := make([][]string, 0, len(values))
	for _, value := range values {
		cells, ok := value.([]any)
		if !ok {
			continue
		}
		row := make([]string, 0, len(cells))
		for _, cell := range cells {
			switch typed := cell.(type) {
			case string:
				row = append(row, typed)
			case json.Number:
				row = append(row, typed.String())
			default:
				row = append(row, "")
			}
		}
		rows = append(rows, row)
	}
	return rows
}

func styleString(node document.Component, key string) string {
	value, _ := node.Style[key].(string)
	return value
}

// overlay renders the Beamer overlay specification carried in Style["overlay"],
// or "" when the node carries none.
func overlay(node document.Component) string {
	spec := strings.TrimSpace(styleString(node, "overlay"))
	if spec == "" {
		return ""
	}
	return "<" + spec + ">"
}

// ── escaping ────────────────────────────────────────────────────────────────

var textEscaper = strings.NewReplacer(
	`\`, `\textbackslash{}`,
	`{`, `\{`,
	`}`, `\}`,
	`$`, `\$`,
	`&`, `\&`,
	`#`, `\#`,
	`_`, `\_`,
	`%`, `\%`,
	`~`, `\textasciitilde{}`,
	`^`, `\textasciicircum{}`,
)

// escape makes arbitrary document text safe as LaTeX body text.
func escape(value string) string { return textEscaper.Replace(value) }

// escapePath makes a path safe inside \includegraphics and \IfFileExists.
func escapePath(value string) string {
	return strings.NewReplacer(`\`, `/`, `{`, ``, `}`, ``, `%`, `\%`, `#`, `\#`).Replace(value)
}

// ── renderers ───────────────────────────────────────────────────────────────

func out(content string, node document.Component) (component.Output, error) {
	return component.Output{Content: []byte(content), Assets: node.Assets}, nil
}

func renderSection(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	var frame strings.Builder

	background := hexColor(styleString(node, "backgroundColor"))
	if background != "" {
		frame.WriteString("{\\definecolor{platoframebg}{HTML}{" + background + "}\n")
		frame.WriteString("\\setbeamercolor{background canvas}{bg=platoframebg}\n")
	}

	frame.WriteString("\\begin{frame}[fragile]\n")
	if title := propString(node, "title"); title != "" {
		frame.WriteString("\\frametitle{" + escape(title) + "}\n")
	}
	frame.WriteString(ChildrenMarker + "\n\\end{frame}\n")
	if background != "" {
		frame.WriteString("}\n")
	}
	return out(frame.String(), node)
}

// hexColor normalises a 3- or 6-digit CSS hex colour to the 6-digit uppercase
// form \definecolor takes, or returns "" for any other value.
func hexColor(value string) string {
	trimmed := strings.TrimPrefix(strings.TrimSpace(value), "#")
	if len(trimmed) == 3 && isHex(trimmed) {
		return strings.ToUpper(string([]byte{
			trimmed[0], trimmed[0], trimmed[1], trimmed[1], trimmed[2], trimmed[2],
		}))
	}
	if len(trimmed) == 6 && isHex(trimmed) {
		return strings.ToUpper(trimmed)
	}
	return ""
}

func isHex(value string) bool {
	for _, character := range value {
		switch {
		case character >= '0' && character <= '9':
		case character >= 'a' && character <= 'f':
		case character >= 'A' && character <= 'F':
		default:
			return false
		}
	}
	return true
}

func renderText(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	// Props["content"] carries raw markdown or html; it is escaped, never
	// interpreted.
	if content := propString(node, "content"); content != "" {
		return out(escape(content)+"\n\n", node)
	}
	return out(ChildrenMarker+"\n\n", node)
}

func renderSpan(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	text := escape(propString(node, "text"))
	for _, mark := range sortedMarks(node) {
		if command, known := MarkCommands[mark]; known {
			text = "\\" + command + "{" + text + "}"
		}
	}
	return out(text, node)
}

func sortedMarks(node document.Component) []string {
	marks := propStrings(node, "marks")
	sort.Strings(marks)
	return marks
}

func renderLink(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	text := escape(propString(node, "text"))
	href := propString(node, "href")
	if href == "" {
		return out(text, node)
	}
	return out("\\href{"+escapePath(href)+"}{"+text+"}", node)
}

func renderHeading(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	// Emitted as sized body text, never as a sectioning command.
	size := map[int]string{1: "\\Large", 2: "\\large", 3: "\\normalsize"}[propInt(node, "level", 3)]
	if size == "" {
		size = "\\normalsize"
	}
	return out("{"+size+"\\bfseries "+ChildrenMarker+"\\par}\n", node)
}

func validateBullets(node document.Component) document.Problems {
	if _, exists := node.Props["items"]; !exists {
		return nil
	}
	if _, ok := node.Props["items"].([]any); !ok {
		return document.Problems{{
			ComponentID: node.ID,
			Path:        "props.items",
			Code:        "component.items.invalid",
			Message:     "items must be an array",
		}}
	}
	return nil
}

func renderBullets(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	environment := "itemize"
	if propBool(node, "ordered") {
		environment = "enumerate"
	}
	// Style["overlay"] wins; Props["fragments"] falls back to one step per item.
	spec := overlay(node)
	if spec == "" && propBool(node, "fragments") {
		spec = "[<+->]"
	} else if spec != "" {
		spec = "[" + spec + "]"
	}
	var list strings.Builder
	list.WriteString("\\begin{" + environment + "}" + spec + "\n")
	for _, item := range propStrings(node, "items") {
		list.WriteString("  \\item " + escape(item) + "\n")
	}
	list.WriteString("\\end{" + environment + "}\n")
	return out(list.String(), node)
}

func renderCode(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	source := propString(node, "source")
	options := []string{}
	if mapped := ListingLanguages[strings.ToLower(strings.TrimSpace(propString(node, "language")))]; mapped != "" {
		options = append(options, "language="+mapped)
	}
	var block strings.Builder
	block.WriteString("\\begin{lstlisting}")
	if len(options) > 0 {
		block.WriteString("[" + strings.Join(options, ",") + "]")
	}
	block.WriteString("\n")
	// The body is verbatim, not escaped; only the terminator is neutralised.
	block.WriteString(strings.ReplaceAll(source, "\\end{lstlisting}", "\\ end{lstlisting}"))
	if !strings.HasSuffix(source, "\n") {
		block.WriteString("\n")
	}
	block.WriteString("\\end{lstlisting}\n")
	return out(block.String(), node)
}

func renderQuote(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	var quote strings.Builder
	quote.WriteString("\\begin{quote}\n")
	quote.WriteString(ChildrenMarker + "\n")
	if cite := propString(node, "cite"); cite != "" {
		quote.WriteString("\\par\\raggedleft\\textemdash\\ " + escape(cite) + "\n")
	}
	quote.WriteString("\\end{quote}\n")
	return out(quote.String(), node)
}

func validateTable(node document.Component) document.Problems {
	rows, exists := node.Props["rows"]
	if !exists {
		return nil
	}
	if _, ok := rows.([]any); !ok {
		return document.Problems{{
			ComponentID: node.ID,
			Path:        "props.rows",
			Code:        "component.rows.invalid",
			Message:     "rows must be an array of arrays",
		}}
	}
	return nil
}

func renderTable(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	head := propStrings(node, "head")
	rows := propRows(node, "rows")
	columns := len(head)
	for _, row := range rows {
		if len(row) > columns {
			columns = len(row)
		}
	}
	if columns == 0 {
		return out("", node)
	}

	var table strings.Builder
	table.WriteString("\\begin{center}\n\\begin{tabular}{" + strings.Repeat("l", columns) + "}\n\\hline\n")
	if len(head) > 0 {
		table.WriteString(row(head, columns, true) + "\\hline\n")
	}
	for _, cells := range rows {
		table.WriteString(row(cells, columns, false))
	}
	table.WriteString("\\hline\n\\end{tabular}\n")
	if caption := propString(node, "caption"); caption != "" {
		table.WriteString("\\par\\smallskip\\footnotesize " + escape(caption) + "\n")
	}
	table.WriteString("\\end{center}\n")
	return out(table.String(), node)
}

// row renders one tabular row, padding to exactly `columns` cells.
func row(cells []string, columns int, header bool) string {
	rendered := make([]string, columns)
	for index := range rendered {
		cell := ""
		if index < len(cells) {
			cell = escape(cells[index])
		}
		if header && cell != "" {
			cell = "\\textbf{" + cell + "}"
		}
		rendered[index] = cell
	}
	return strings.Join(rendered, " & ") + " \\\\\n"
}

// unprintable reports why a source may not reach \includegraphics, or "" when
// it may. Callers must render a placeholder for any non-empty reason.
func unprintable(source string) string {
	extension := strings.ToLower(pathExt(source))
	if PrintableExtensions[extension] {
		return ""
	}
	if reason, known := UnprintableReasons[extension]; known {
		return reason
	}
	if extension == "" {
		return "no file extension"
	}
	return "unsupported format " + extension
}

func pathExt(value string) string {
	if index := strings.LastIndex(value, "."); index >= 0 && index > strings.LastIndex(value, "/") {
		return value[index:]
	}
	return ""
}

// placeholder renders a framed box naming the label, caption and source. It is
// the required rendering for any content print cannot show.
func placeholder(label, source, caption string) string {
	var block strings.Builder
	block.WriteString("\\begin{center}\n\\fbox{\\parbox{0.8\\textwidth}{\\centering\\footnotesize\n")
	block.WriteString("\\textbf{" + escape(strings.ToUpper(label)) + "}")
	if caption != "" {
		block.WriteString(" \\textemdash\\ " + escape(caption))
	}
	if source != "" {
		block.WriteString("\\par\\ttfamily " + escape(source))
	}
	block.WriteString("\n}}\n\\end{center}\n")
	return block.String()
}

// figure emits an \includegraphics guarded by \IfFileExists, falling back to a
// visible "missing" box.
func figure(source, options, caption string) string {
	path := escapePath(source)
	var block strings.Builder
	block.WriteString("\\begin{center}\n")
	block.WriteString("\\IfFileExists{" + path + "}%\n")
	block.WriteString("  {\\includegraphics[" + options + "]{" + path + "}}%\n")
	block.WriteString("  {\\fbox{\\ttfamily missing: " + escape(source) + "}}\n")
	if caption != "" {
		block.WriteString("\\par\\smallskip\\footnotesize " + escape(caption) + "\n")
	}
	block.WriteString("\\end{center}\n")
	return block.String()
}

func renderImage(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	source := propString(node, "src")
	if source == "" {
		return out("", node)
	}
	caption := propString(node, "caption")
	if reason := unprintable(source); reason != "" {
		return out(placeholder(reason, source, caption), node)
	}
	options := "width=0.8\\textwidth,keepaspectratio"
	if width := propString(node, "width"); width != "" {
		options = "width=" + escape(width) + ",keepaspectratio"
	}
	return out(figure(source, options, caption), node)
}

func renderColumns(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	return out("\\begin{columns}[T]\n"+ChildrenMarker+"\n\\end{columns}\n", node)
}

func renderCards(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	return out(ChildrenMarker+"\n", node)
}

func renderCard(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	var card strings.Builder
	card.WriteString("\\begin{tcolorbox}[")
	if title := propString(node, "title"); title != "" {
		card.WriteString("title=" + escape(title))
	} else {
		card.WriteString("blanker,boxrule=0.4pt,colframe=gray")
	}
	card.WriteString("]\n" + ChildrenMarker + "\n\\end{tcolorbox}\n")
	return out(card.String(), node)
}

func renderCallout(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	tone, known := CalloutTones[propString(node, "tone")]
	if !known {
		tone = CalloutTones[DefaultTone]
	}
	options := "colback=" + tone.Background + ",colframe=" + tone.Frame
	if title := propString(node, "title"); title != "" {
		options += ",title=" + escape(title)
	}
	return out("\\begin{tcolorbox}["+options+"]\n"+ChildrenMarker+"\n\\end{tcolorbox}\n", node)
}

func renderKicker(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	return out("{\\footnotesize\\scshape "+escape(propString(node, "text"))+"\\par}\\smallskip\n", node)
}

// renderMedia emits a printable poster when there is one, followed by a
// placeholder naming the medium, caption and source. It never emits \movie or
// media9.
func renderMedia(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	mediaType := propString(node, "mediaType")
	if mediaType == "" {
		mediaType = "media"
	}
	var block strings.Builder
	if poster := propString(node, "poster"); poster != "" && unprintable(poster) == "" {
		path := escapePath(poster)
		block.WriteString("\\begin{center}\n")
		block.WriteString("\\IfFileExists{" + path + "}%\n")
		block.WriteString("  {\\includegraphics[width=0.7\\textwidth,keepaspectratio]{" + path + "}}%\n")
		block.WriteString("  {}%\n")
		block.WriteString("\\end{center}\n")
	}
	block.WriteString(placeholder(mediaType, propString(node, "src"), propString(node, "caption")))
	return out(block.String(), node)
}

func renderNotes(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	text := propString(node, "text")
	if strings.TrimSpace(text) == "" {
		return out("", node)
	}
	return out("\\note{"+escape(text)+"}\n", node)
}

func renderRule(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	return out("\\par\\noindent\\rule{\\textwidth}{0.4pt}\\par\n", node)
}

func renderScene(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	// Props["svg"] is a snapshot path; it prints only once the asset boundary
	// has converted it to a PrintableExtensions format.
	source := propString(node, "svg")
	if source == "" {
		return out("", node)
	}
	if reason := unprintable(source); reason != "" {
		return out(placeholder("scene — "+reason, source, propString(node, "caption")), node)
	}
	return out(figure(source, "width=0.8\\textwidth,keepaspectratio", propString(node, "caption")), node)
}
