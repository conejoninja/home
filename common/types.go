package common

import (
	"time"
	"reflect"
	"fmt"
)

type Device struct {
	Id      string   `json:"id"`
	Name    string   `json:"name,omitempty"`
	Version string   `json:"version,omitempty"`
	Out     []Value  `json:"out,omitempty"`
	Methods []Method `json:"methods,omitempty"`
}

type Value struct {
	Id    string      `json:"id"`
	Type  string      `json:"type,omitempty"`
	Name  string      `json:"name,omitempty"`
	Unit  string      `json:"unit,omitempty"`
	Min   string      `json:"min,omitempty"`
	Max   string      `json:"max,omitempty"`
	Time  *time.Time   `json:"time,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

type Method struct {
	Name   string  `json:"name"`
	Params []Value `json:"params,omitempty"`
}

type Param struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Event struct {
	Id       string    `json:"id"`
	Message  string    `json:"message,omitempty"`
	Priority uint8     `json:"priority,omitempty"`
	Time     *time.Time `json:"time,omitempty"`
	Extra    []Param   `json:"extra,omitempty"`
}

type Meta struct {
	Max float64 `json:"max,omitempty"`
	Min float64 `json:"min,omitempty"`
	Avg float64 `json:"avg,omitempty"`
	N   int     `json:"n,omitempty"`
}


var floatType = reflect.TypeOf(float64(0))
func GetFloat(unk interface{}) (float64, error) {
	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	if unk == nil {
		return 0, nil
	}
	if !v.Type().ConvertibleTo(floatType) {
		return 0, fmt.Errorf("cannot convert %v to float64", v.Type())
	}
	fv := v.Convert(floatType)
	return fv.Float(), nil
}

func (d *Device) IsNil() bool {
	if d.Id != "" ||
		d.Name != "" {
		return false
	}

	return true
}
