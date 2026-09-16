// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package strict

import (
	"strings"
	"text/template"
	"text/template/parse"
)

// References collects every root-scoped variable read in a parsed template.
//
// source must be the text the template was parsed from; it is used only to
// turn byte offsets into line and column numbers.
//
// Reads whose dot has been rebound (the body of a range or a with) are NOT
// collected, because their key is relative to a value this function cannot
// know. Those reads are held strict at execution time by the template option
// missingkey=error instead.
func References(tmpl *template.Template, source string) []Reference {
	if tmpl == nil || tmpl.Tree == nil || tmpl.Tree.Root == nil {
		return nil
	}
	w := &walker{source: source}
	w.node(tmpl.Tree.Root, false)
	return w.refs
}

type walker struct {
	source string
	refs   []Reference
}

// waiver is what a site's `optional` call says about one argument node.
type waiver struct {
	reason string
	// handled reports that the pipe already accounted for this node, so the
	// ordinary walk must not record it a second time.
	handled bool
}

type waivers map[parse.Node]waiver

func (w *walker) node(n parse.Node, scoped bool) {
	switch typed := n.(type) {
	case nil:
		return
	case *parse.ListNode:
		if typed == nil {
			return
		}
		for _, child := range typed.Nodes {
			w.node(child, scoped)
		}
	case *parse.ActionNode:
		w.pipe(typed.Pipe, scoped)
	case *parse.IfNode:
		// if does not rebind the dot: the whole branch stays in scope.
		w.branch(&typed.BranchNode, scoped, scoped)
	case *parse.RangeNode:
		// range rebinds the dot in its body; its else branch does not.
		w.branch(&typed.BranchNode, true, scoped)
	case *parse.WithNode:
		// with rebinds the dot in its body; its else branch does not.
		w.branch(&typed.BranchNode, true, scoped)
	case *parse.TemplateNode:
		w.pipe(typed.Pipe, scoped)
	}
}

func (w *walker) branch(b *parse.BranchNode, bodyScoped, elseScoped bool) {
	w.pipe(b.Pipe, elseScoped)
	if b.List != nil {
		w.node(b.List, bodyScoped)
	}
	if b.ElseList != nil {
		w.node(b.ElseList, elseScoped)
	}
}

func (w *walker) pipe(p *parse.PipeNode, scoped bool) {
	if p == nil {
		return
	}
	waived := waivers{}
	for index, cmd := range p.Cmds {
		if !isWaiverCall(cmd) {
			continue
		}
		target, reason, ok := waiverTarget(index, cmd)
		if !ok {
			w.recordMalformed(cmd, target)
			if target != nil {
				waived[target] = waiver{handled: true}
			}
			continue
		}
		waived[target] = waiver{reason: reason}
	}
	for _, cmd := range p.Cmds {
		w.command(cmd, scoped, waived)
	}
}

func (w *walker) command(cmd *parse.CommandNode, scoped bool, waived waivers) {
	if cmd == nil {
		return
	}
	for _, arg := range cmd.Args {
		w.arg(arg, scoped, waived)
	}
}

func (w *walker) arg(arg parse.Node, scoped bool, waived waivers) {
	if declared, isWaived := waived[arg]; isWaived {
		if declared.handled {
			return
		}
		if key, ok := readKey(arg); ok && !scoped {
			w.record(arg, key, true, declared.reason)
		}
		return
	}

	switch typed := arg.(type) {
	case *parse.FieldNode:
		if !scoped {
			w.record(arg, strings.Join(typed.Ident, "."), false, "")
		}
	case *parse.VariableNode:
		// $ is the root data whatever the dot currently is.
		if len(typed.Ident) > 1 && typed.Ident[0] == "$" {
			w.record(arg, strings.Join(typed.Ident[1:], "."), false, "")
		}
	case *parse.PipeNode:
		w.pipe(typed, scoped)
	case *parse.ChainNode:
		// A field read off a computed value: the container is not the root.
		w.arg(typed.Node, scoped, waived)
	}
}

func (w *walker) record(at parse.Node, key string, isWaived bool, reason string) {
	line, column := position(w.source, at.Position())
	w.refs = append(w.refs, Reference{
		Key:    key,
		Line:   line,
		Column: column,
		Waived: isWaived,
		Reason: reason,
	})
}

func isWaiverCall(cmd *parse.CommandNode) bool {
	if cmd == nil || len(cmd.Args) == 0 {
		return false
	}
	ident, isIdent := cmd.Args[0].(*parse.IdentifierNode)
	return isIdent && ident.Ident == FuncName
}

// waiverTarget reads the command form `optional .key "reason"`. Any other
// shape, the pipe form included, is refused: a pipe would hand optional its
// arguments in the opposite order and print the reason instead of the value.
func waiverTarget(index int, cmd *parse.CommandNode) (target parse.Node, reason string, ok bool) {
	if index != 0 || len(cmd.Args) < 2 {
		return nil, "", false
	}
	target = cmd.Args[1]
	if _, isRead := readKey(target); !isRead {
		return target, "", false
	}
	if len(cmd.Args) != 3 {
		return target, "", false
	}
	str, isString := cmd.Args[2].(*parse.StringNode)
	if !isString {
		return target, "", false
	}
	return target, str.Text, true
}

func (w *walker) recordMalformed(cmd *parse.CommandNode, target parse.Node) {
	key := ""
	if target != nil {
		if read, ok := readKey(target); ok {
			key = read
		}
	}
	line, column := position(w.source, cmd.Position())
	w.refs = append(w.refs, Reference{
		Key:       key,
		Line:      line,
		Column:    column,
		Waived:    true,
		Malformed: true,
	})
}

// readKey returns the dotted key a node reads off the root data.
func readKey(n parse.Node) (string, bool) {
	switch typed := n.(type) {
	case *parse.FieldNode:
		return strings.Join(typed.Ident, "."), true
	case *parse.VariableNode:
		if len(typed.Ident) > 1 && typed.Ident[0] == "$" {
			return strings.Join(typed.Ident[1:], "."), true
		}
	}
	return "", false
}

// position converts a byte offset into a 1-based line and column.
func position(source string, pos parse.Pos) (line, column int) {
	offset := int(pos)
	if offset < 0 {
		offset = 0
	}
	if offset > len(source) {
		offset = len(source)
	}
	head := source[:offset]
	return 1 + strings.Count(head, "\n"), offset - strings.LastIndexByte(head, '\n')
}
