// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package parser

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/ta2gch/iris/reader/tokenizer"
	"github.com/ta2gch/iris/runtime/ilos"
	"github.com/ta2gch/iris/runtime/ilos/class"
	"github.com/ta2gch/iris/runtime/ilos/instance"
)

var eop = instance.Symbol("End Of Parentheses")
var bod = instance.Symbol("Begin Of Dot")

func ParseAtom(tok string) (ilos.Instance, ilos.Instance) {
	//
	// integer
	//
	if m, _ := regexp.MatchString("^[-+]?[[:digit:]]+$", tok); m {
		n, _ := strconv.ParseInt(tok, 10, 64)
		return instance.Integer(int(n)), nil
	}
	if r := regexp.MustCompile("^#[bB]([-+]?[01]+)$").FindStringSubmatch(tok); len(r) >= 2 {
		n, _ := strconv.ParseInt(r[1], 2, 64)
		return instance.Integer(int(n)), nil
	}
	if r := regexp.MustCompile("^#[oO]([-+]?[0-7]+)$").FindStringSubmatch(tok); len(r) >= 2 {
		n, _ := strconv.ParseInt(r[1], 8, 64)
		return instance.Integer(int(n)), nil
	}
	if r := regexp.MustCompile("^#[xX]([-+]?[[:xdigit:]]+)$").FindStringSubmatch(tok); len(r) >= 2 {
		n, _ := strconv.ParseInt(r[1], 16, 64)
		return instance.Integer(int(n)), nil
	}
	//
	// float
	//
	if m, _ := regexp.MatchString("^[-+]?[[:digit:]]+\\.[[:digit:]]+$", tok); m {
		n, _ := strconv.ParseFloat(tok, 64)
		return instance.Float(n), nil
	}
	if r := regexp.MustCompile("^([-+]?[[:digit:]]+(?:\\.[[:digit:]]+)?)[eE]([-+]?[[:digit:]]+)$").FindStringSubmatch(tok); len(r) >= 3 {
		n, _ := strconv.ParseFloat(r[1], 64)
		e, _ := strconv.ParseInt(r[2], 10, 64)
		return instance.Float(n * math.Pow10(int(e))), nil
	}
	//
	// character
	//
	if m, _ := regexp.MatchString("^#\\\\newline$", tok); m {
		return instance.Character('\n'), nil
	}
	if m, _ := regexp.MatchString("^#\\\\space$", tok); m {
		return instance.Character(' '), nil
	}
	if r := regexp.MustCompile("^#\\\\([[:graph:]])$").FindStringSubmatch(tok); len(r) >= 2 {
		return instance.Character(rune(r[1][0])), nil
	}
	//
	// string
	//
	if m, _ := regexp.MatchString("^\".*\"$", tok); m {
		return instance.String(tok), nil
	}
	//
	// symbol
	//
	if "nil" == tok {
		return instance.New(class.Null), nil
	}
	str := `^(`
	str += `[:&][a-zA-Z]+|`
	str += `\|.*\||`
	str += `\+|\-|1\+|1\-|`
	str += `[a-zA-Z<>/*=?_!$%[\]^{}~][-a-zA-Z0-9+<>/*=?_!$%[\]^{}~]*|`
	str += ")$"
	if m, _ := regexp.MatchString(str, tok); m {
		return instance.Symbol(strings.ToUpper(tok)), nil
	}
	return nil, instance.New(class.ParseError, map[string]ilos.Instance{
		"STRING":         instance.String(tok),
		"EXPECTED-CLASS": class.Object,
	})
}

func parseMacro(tok string, t *tokenizer.Tokenizer) (ilos.Instance, ilos.Instance) {
	cdr, err := Parse(t)
	if err != nil {
		return nil, err
	}
	n := tok
	if m, _ := regexp.MatchString("#[[:digit:]]*[aA]", tok); m {
		s := instance.Symbol("array")
		i := strings.IndexRune(strings.ToLower(tok), 'a')
		if i == 1 {
			d := instance.Integer(1)
			return instance.New(class.Cons, s, instance.New(class.Cons, d, instance.New(class.Cons, cdr, instance.New(class.Null)))), nil
		}
		v, err := strconv.ParseInt(tok[1:i], 10, 32)
		if err != nil {
			return nil, instance.New(class.ParseError, instance.String(tok), class.Integer)
		}
		d := instance.Integer(int(v))
		return instance.New(class.Cons, s, instance.New(class.Cons, d, instance.New(class.Cons, cdr, instance.New(class.Null)))), nil
	}
	switch tok {
	case "#'":
		n = "FUNCTION"
	case ",@":
		n = "commaat"
	case ",":
		n = "comma"
	case "'":
		n = "QUOTE"
	case "`":
		n = "BACKQUOTE"
	}
	m := instance.Symbol(n)
	return instance.New(class.Cons, m, instance.New(class.Cons, cdr, instance.New(class.Null))), nil
}
func parseCons(t *tokenizer.Tokenizer) (ilos.Instance, ilos.Instance) {
	car, err := Parse(t)
	if err == eop {
		return instance.New(class.Null), nil
	}
	if err == bod {
		cdr, err := Parse(t)
		if err != nil {
			return nil, err
		}
		if _, err := Parse(t); err != eop {
			return nil, err
		}
		return cdr, nil
	}
	if err != nil {
		return nil, err
	}
	cdr, err := parseCons(t)
	if err != nil {
		return nil, err
	}
	return instance.New(class.Cons, car, cdr), nil
}

// Parse builds a internal expression from tokens
func Parse(t *tokenizer.Tokenizer) (ilos.Instance, ilos.Instance) {
	tok, err := t.Next()
	if err != nil {
		return nil, err
	}
	if tok == "(" {
		cons, err := parseCons(t)
		if err != nil {
			return nil, err
		}
		return cons, err
	}
	if tok == ")" {
		return nil, eop
	}
	if tok == "." {
		return nil, bod
	}
	if mat, _ := regexp.MatchString("^(?:#'|,@?|'|`|#[[:digit:]]*[aA])$", tok); mat {
		m, err := parseMacro(tok, t)
		if err != nil {
			return nil, err
		}
		return m, nil
	}
	atom, err := ParseAtom(tok)
	if err != nil {
		return nil, err
	}
	return atom, nil
}
