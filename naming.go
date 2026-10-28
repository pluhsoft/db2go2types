package db2go2types

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// goName converts a PostgreSQL name to an exported Go name: "id_user" → "IdUser",
// "in-progress" → "InProgress". Characters that cannot be in an identifier
// separate words. A name starting with a digit gets the prefix "X".
func goName(name string) string {
	s := words(name)
	if s == "" {
		return "X"
	}
	if r, _ := utf8.DecodeRuneInString(s); unicode.IsDigit(r) {
		s = "X" + s
	}
	return s
}

// words joins the letter and digit runs of s, each starting with an upper-case letter.
func words(s string) string {
	var b strings.Builder
	upper := true
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			upper = true
			continue
		}
		if upper {
			r = unicode.ToUpper(r)
			upper = false
		}
		b.WriteRune(r)
	}
	return b.String()
}

// enumConstName is the Go constant of an enum value: PostStatus + "draft" → PostStatusDraft.
// names maps values to custom names, e.g. "#A6D2FF" → "Blue".
func enumConstName(enum, value string, names map[string]string) string {
	if name, ok := names[value]; ok {
		value = name
	}
	suffix := words(value)
	if suffix == "" {
		suffix = "Empty"
	}
	return enum + suffix
}
