package meta

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

//
// Float64
//

type Float64 struct {
	Val float64
	Nullity
	Presence
}

type FloatOptions struct {
	Required     bool
	DiscardBlank bool
	Null         bool
	MinPresent   bool
	Min          float64
	MaxPresent   bool
	Max          float64
	In           []float64
}

func NewFloat64(f float64) Float64 {
	return Float64{f, Nullity{false}, Presence{true}}
}

func (i *Float64) ParseOptions(tag reflect.StructTag) interface{} {
	opts := &FloatOptions{
		DiscardBlank: true,
	}

	if tag.Get("meta_required") == "true" {
		opts.Required = true
	}

	if tag.Get("meta_discard_blank") == "false" {
		opts.DiscardBlank = false
	}

	if tag.Get("meta_null") == "true" {
		opts.Null = true
	}

	if nstr := tag.Get("meta_min"); nstr != "" {
		n, err := strconv.ParseFloat(nstr, 64)
		if err != nil {
			panic(err.Error())
		}

		opts.MinPresent = true
		opts.Min = n
	}

	if nstr := tag.Get("meta_max"); nstr != "" {
		n, err := strconv.ParseFloat(nstr, 64)
		if err != nil {
			panic(err.Error())
		}

		opts.MaxPresent = true
		opts.Max = n
	}

	if in := tag.Get("meta_in"); in != "" {
		for _, s := range strings.Split(in, ",") {
			n, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
			if err != nil {
				panic(err.Error())
			}

			opts.In = append(opts.In, n)
		}
	}

	return opts
}

func (f *Float64) JSONValue(i interface{}, options interface{}) Errorable {
	if i == nil {
		return f.FormValue("", options)
	}

	switch value := i.(type) {
	case json.Number:
		return f.FormValue(string(value), options)
	case string:
		return f.FormValue(string(value), options)
	}
	return ErrFloat
}

func (i *Float64) FormValue(value string, options interface{}) Errorable {
	opts := options.(*FloatOptions)

	if value == "" {
		if opts.Null {
			i.Null = true
			i.Present = true
			return nil
		}
		if opts.Required {
			return ErrBlank
		}
		if !opts.DiscardBlank {
			i.Present = true
			return ErrBlank
		}
		return nil
	}

	if n, err := strconv.ParseFloat(value, 64); err == nil {
		if opts.MinPresent && n < opts.Min {
			return ErrMin
		}
		if opts.MaxPresent && n > opts.Max {
			return ErrMax
		}
		if len(opts.In) > 0 {
			found := false
			for _, i := range opts.In {
				if i == n {
					found = true
				}
			}
			if !found {
				return ErrIn
			}
		}

		i.Val = n
		i.Present = true
	} else {
		numError := err.(*strconv.NumError)
		if numError.Err == strconv.ErrRange {
			return ErrFloatRange
		} else {
			return ErrFloat
		}
	}
	return nil
}

func (i Float64) Value() (driver.Value, error) {
	if i.Present && !i.Null {
		return float64(i.Val), nil
	}
	return nil, nil
}

func (i Float64) MarshalJSON() ([]byte, error) {
	if i.Present && !i.Null {
		return json.Marshal(i.Val)
	}
	return nullString, nil
}
