package main

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/k0kubun/pp"
	"github.com/spf13/pflag"
)

var Format format = textFormat{}

func registerFormatFlags(flags *pflag.FlagSet) {
	flags.VarP(formatVar{value: &Format}, "format", "f", "The format use when printing output")
}

type format interface {
	Identifier() string
	Print(w io.Writer, value any) error
}

type formatVar struct {
	value *format
}

func (f formatVar) String() string { return (*f.value).Identifier() }

func (f formatVar) Set(value string) error {
	switch value {
	case "text":
		*f.value = textFormat{}
		return nil
	case "json":
		*f.value = jsonFormat{}
		return nil
	default:
		return errors.New("unknown format")
	}
}

func (f formatVar) Type() string { return "string" }

type jsonFormat struct{}

func (t jsonFormat) Identifier() string {
	return "json"
}

func (t jsonFormat) Print(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

type textFormat struct{}

func (t textFormat) Identifier() string {
	return "text"
}

func (t textFormat) Print(w io.Writer, value any) error {
	_, err := pp.Fprintln(w, value)
	return err
}
