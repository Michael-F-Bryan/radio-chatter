package main

import "errors"

type Format string

const (
	TextFormat Format = "text"
	JsonFormat Format = "json"
)

func (f Format) String() string { return string(f) }

func (f *Format) Set(value string) error {
	switch value {
	case "text":
		*f = TextFormat
		return nil
	case "json":
		*f = JsonFormat
		return nil
	default:
		return errors.New("unknown format")
	}
}

func (f Format) Type() string { return "string" }
