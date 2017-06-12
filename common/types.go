package common

import "time"

type Device struct {
	Id      string   `json:"id"`
	Name    string   `json:"name,omitempty"`
	Version string   `json:"version,omitempty"`
	Out     []Value  `json:"out,omitempty"`
	Methods []Method `json:"methods,omitempty"`
}

type Value struct {
	Id    string      `json:"id"`
	Type  string      `json:"type"`
	Name  string      `json:"name,omitempty"`
	Unit  string      `json:"unit,omitempty"`
	Min   string      `json:"min,omitempty"`
	Max   string      `json:"max,omitempty"`
	Time  time.Time   `json:"time,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

type Method struct {
	Name   string  `json:"name"`
	Params []Param `json:"params,omitempty"`
}

type Param struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (d *Device) IsNil() bool {
	if d.Id != "" ||
		d.Name != "" {
		return false
	}

	return true
}
