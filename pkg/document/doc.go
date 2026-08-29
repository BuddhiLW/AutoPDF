// Package document defines AutoPDF's inert, versioned document model.
//
// A DocumentSpec is the source of truth for composition. Renderers may project
// it into LaTeX, HTML, or another target, but generated fragments are never
// written back into the domain model.
package document
