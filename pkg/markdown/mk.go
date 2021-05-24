package markdown

import (
	"fmt"
	"strings"
)

// Separate is a horizontal rule <hr />.
const Separate = "<hr />" //"---------"

// Escape backslash escapes strings in markdown.
//
// Backslash Escapes
//
// Markdown allows you to use backslash escapes to generate literal characters
// which would otherwise have special meaning in Markdown’s formatting syntax.
// For example, if you wanted to surround a word with literal asterisks
// (instead of an HTML <em> tag), you can backslashes before the asterisks,
// like this:
//
//   \*literal asterisks\*
//
// Markdown provides backslash escapes for the following characters:
//
//   \   backslash
//   `   backtick
//   *   asterisk
//   _   underscore
//   {}  curly braces
//   []  square brackets
//   ()  parentheses
//   #   hash mark
//   +   plus sign
//   -   minus sign (hyphen)
//   .   dot
//   !   exclamation mark
//
// Source: https://golem.ph.utexas.edu/~distler/maruku/markdown_syntax.html#backslash
func Escape(input string) string {
	for _, token := range []string{
		"\\",
		"`",
		"*",
		"_",
		"{", "}",
		"[", "]",
		"(", ")",
		"#",
		"+",
		"-",
		".",
		"!",
	} {
		input = strings.ReplaceAll(input, token, "\\"+token)
	}
	return input
}

// Link builds a inline style markdown link.
func Link(identifier string, url string) string {
	return fmt.Sprintf("[%s](%s)", Escape(identifier), url)
}
