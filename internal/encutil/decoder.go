// Copyright 2022 Franklin "Snaipe" Mathieu.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

package encutil

import (
	"fmt"
	"io"
	"io/fs"
	"reflect"

	"snai.pe/boa/encoding"
	"snai.pe/boa/internal/reflectutil"
	"snai.pe/boa/syntax"
)

type UnmarshalerBase struct {
	NewParser func(io.Reader) syntax.Parser
	Self      reflectutil.Unmarshaler

	encoding.CommonOptions
	encoding.DecoderOptions
}

func (unmarshaler *UnmarshalerBase) Decode(in io.Reader, v interface{}) error {
	ptr := reflect.ValueOf(v)
	if ptr.Kind() != reflect.Ptr {
		panic("decode: must pass in pointer value")
	}

	root, err := unmarshaler.NewParser(in).Parse()
	if err != nil {
		if e, ok := err.(*syntax.Error); ok {
			e.Filename = Name(in)
		}
		return err
	}
	if node, ok := v.(**syntax.Node); ok {
		*node = root
		return nil
	}
	err = reflectutil.Unmarshal(ptr.Elem(), root.Child, unmarshaler.NamingConvention, unmarshaler.Self)
	if e, ok := err.(*encoding.LoadError); ok {
		e.Filename = Name(in)
	}
	return err
}

func (unmarshaler *UnmarshalerBase) Option(opts ...interface{}) error {
	for _, opt := range opts {
		switch setopt := opt.(type) {
		case encoding.CommonOption:
			setopt(&unmarshaler.CommonOptions)
		case encoding.DecoderOption:
			setopt(&unmarshaler.DecoderOptions)
		default:
			return fmt.Errorf("%T is not a common option, nor an decoder option.", opt)
		}
	}
	return nil
}

type Stater interface {
	Stat() (fs.FileInfo, error)
}

type Namer interface {
	Name() string
}

func Name(in interface{}) string {
	if namer, ok := in.(Namer); ok {
		return namer.Name()
	}
	if stater, ok := in.(Stater); ok {
		if info, err := stater.Stat(); err == nil {
			return info.Name()
		}
	}
	return ""
}
