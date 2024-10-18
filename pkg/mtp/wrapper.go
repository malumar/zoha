package mtp

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Godoc comment extraction and comment -> HTML formatting.

// original go/doc/comment.go

import (
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

func indentLen(s string) int {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	return i
}

func isBlank(s string) bool {
	return len(s) == 0 || (len(s) == 1 && s[0] == '\n')
}

func commonPrefix(a, b string) string {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	return a[0:i]
}

func unindent(block []string) {
	if len(block) == 0 {
		return
	}

	// compute maximum common white prefix
	prefix := block[0][0:indentLen(block[0])]
	for _, line := range block {
		if !isBlank(line) {
			prefix = commonPrefix(prefix, line[0:indentLen(line)])
		}
	}
	n := len(prefix)

	// remove
	for i, line := range block {
		if !isBlank(line) {
			block[i] = line[n:]
		}
	}
}

// heading returns the trimmed line if it passes as a section heading;
// otherwise it returns the empty string.
func heading(line string) string {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return ""
	}

	// a heading must start with an uppercase letter
	r, _ := utf8.DecodeRuneInString(line)
	if !unicode.IsLetter(r) || !unicode.IsUpper(r) {
		return ""
	}

	// it must end in a letter or digit:
	r, _ = utf8.DecodeLastRuneInString(line)
	if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
		return ""
	}

	// exclude lines with illegal characters
	if strings.IndexAny(line, ",.;:!?+*/=()[]{}_^°&§~%#@<\">\\") >= 0 {
		return ""
	}

	// allow "'" for possessive "'s" only
	for b := line; ; {
		i := strings.IndexRune(b, '\'')
		if i < 0 {
			break
		}
		if i+1 >= len(b) || b[i+1] != 's' || (i+2 < len(b) && b[i+2] != ' ') {
			return "" // not followed by "s "
		}
		b = b[i+2:]
	}

	return line
}

type op int

const (
	opPara op = iota
	opHead
	opPre
)

type block struct {
	op    op
	lines []string
}

func blocks(text string) []block {
	var (
		out  []block
		para []string

		lastWasBlank   = false
		lastWasHeading = false
	)

	close := func() {
		if para != nil {
			out = append(out, block{opPara, para})
			para = nil
		}
	}

	lines := strings.SplitAfter(text, "\n")
	unindent(lines)
	for i := 0; i < len(lines); {
		line := lines[i]
		if isBlank(line) {
			// close paragraph
			close()
			i++
			lastWasBlank = true
			continue
		}
		if indentLen(line) > 0 {
			// close paragraph
			close()

			// count indented or blank lines
			j := i + 1
			for j < len(lines) && (isBlank(lines[j]) || indentLen(lines[j]) > 0) {
				j++
			}
			// but not trailing blank lines
			for j > i && isBlank(lines[j-1]) {
				j--
			}
			pre := lines[i:j]
			i = j

			unindent(pre)

			// put those lines in a pre block
			out = append(out, block{opPre, pre})
			lastWasHeading = false
			continue
		}

		if lastWasBlank && !lastWasHeading && i+2 < len(lines) &&
			isBlank(lines[i+1]) && !isBlank(lines[i+2]) && indentLen(lines[i+2]) == 0 {
			// current line is non-blank, surrounded by blank lines
			// and the next non-blank line is not indented: this
			// might be a heading.
			if head := heading(line); head != "" {
				close()
				out = append(out, block{opHead, []string{head}})
				i += 2
				lastWasHeading = true
				continue
			}
		}

		// open paragraph
		lastWasBlank = false
		lastWasHeading = false
		para = append(para, lines[i])
		i++
	}
	close()

	return out
}

// WrapText prepares comment text for presentation in textual output.
// It wraps paragraphs of text to width or fewer Unicode code points
// and then prefixes each line with the indent.  In preformatted sections
// (such as program text), it prefixes each non-blank line with preIndent.
// newline - lamacz lini
func WrapText(w io.Writer, text string, newline, indent, preIndent string, width int) {
	l := lineWrapper{
		out:    w,
		width:  width,
		indent: indent,
	}
	for _, b := range blocks(text) {
		switch b.op {
		case opPara:
			// l.write will add leading newline if required
			for _, line := range b.lines {
				l.write(line)
			}
			l.flush()
		case opHead:
			w.Write(nl)

			for _, line := range b.lines {
				l.write(line + "\n")
			}
			l.flush()
		case opPre:
			w.Write(nl)

			for _, line := range b.lines {
				if !isBlank(line) {
					w.Write([]byte(preIndent))
					w.Write([]byte(line))
				}
			}
		}
	}
}

type lineWrapper struct {
	out       io.Writer
	printed   bool
	width     int
	indent    string
	n         int
	pendSpace int
}

var nl = []byte("\n")
var space = []byte(" ")

// bo outlooki szaleja
var tabWithOutCr = []byte("\n\t")

func (l *lineWrapper) write(text string) {
	if l.n == 0 && l.printed {
		l.out.Write(nl) // blank line before new paragraph
	}
	l.printed = true

	for _, f := range strings.Fields(text) {
		w := utf8.RuneCountInString(f)
		// wrap if line is too long
		if l.n > 0 && l.n+l.pendSpace+w > l.width {
			l.out.Write(tabWithOutCr)
			//l.out.Write(nl)
			l.n = 0
			l.pendSpace = 0
		}
		if l.n == 0 {
			l.out.Write([]byte(l.indent))
		}
		l.out.Write(space[:l.pendSpace])
		l.out.Write([]byte(f))
		l.n += l.pendSpace + w
		l.pendSpace = 1
	}
}

func (l *lineWrapper) flush() {
	if l.n == 0 {
		return
	}
	l.out.Write(nl)
	l.pendSpace = 0
	l.n = 0
}
