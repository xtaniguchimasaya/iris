// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package runtime

import (
	"strings"

	"github.com/ta2gch/iris/runtime/environment"

	"github.com/ta2gch/iris/reader/parser"
	"github.com/ta2gch/iris/reader/tokenizer"
	"github.com/ta2gch/iris/runtime/ilos"
	"github.com/ta2gch/iris/runtime/ilos/class"
	"github.com/ta2gch/iris/runtime/ilos/instance"
)

// IsProperList tests given argument is proper list
// Proper list is the list terminated by nil.
func IsProperList(i ilos.Instance) bool {
	if instance.Of(class.Cons, i) {
		return IsProperList(instance.UnsafeCdr(i))
	}
	if instance.Of(class.Null, i) {
		return true
	}
	return false
}

func readFromString(s string) ilos.Instance {
	e, _ := parser.Parse(tokenizer.New(strings.NewReader(s)))
	return e
}

func defspecial(name string, macro interface{}) {
	symbol := instance.New(class.Symbol, name)
	environment.TopLevel.Special.Define(symbol, instance.New(class.Function, symbol, macro))
}

func defmacro(name string, macro interface{}) {
	symbol := instance.New(class.Symbol, name)
	environment.TopLevel.Macro.Define(symbol, instance.New(class.Function, symbol, macro))
}

func defun(name string, function interface{}) {
	symbol := instance.New(class.Symbol, name)
	environment.TopLevel.Function.Define(symbol, instance.New(class.Function, symbol, function))
}
func defglobal(name string, value ilos.Instance) {
	symbol := instance.New(class.Symbol, name)
	environment.TopLevel.Variable.Define(symbol, value)
}
