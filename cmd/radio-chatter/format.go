package main

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/k0kubun/pp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func registerFormatFlags(flags *pflag.FlagSet) {
	flags.StringP("format", "f", "text", "The format use when printing output")
	_ = viper.BindPFlag("out.format", flags.Lookup("format"))
	viper.BindEnv("out.format", "FORMAT")
}

type formatter interface {
	Print(w io.Writer, value any) error
}

func getFormatter(name string) (formatter, error) {
	switch name {
	case "text":
		return formatFunc(textFormat), nil
	case "json":

		return formatFunc(jsonFormat), nil
	default:
		return nil, errors.New("unknown format")
	}
}

type formatFunc func(w io.Writer, value any) error

func (f formatFunc) Print(w io.Writer, value any) error {
	return f(w, value)
}

func jsonFormat(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func textFormat(w io.Writer, value any) error {
	_, err := pp.Fprintln(w, value)
	return err
}
