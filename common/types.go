package common

import (
	"fmt"
	"reflect"
	"time"
)

// Device type: Description of the device, which sensor data it sends and which methods it has
type Device struct {
	ID      string   `json:"id"`
	Name    string   `json:"name,omitempty"`
	Version string   `json:"version,omitempty"`
	Out     []Value  `json:"out,omitempty"`
	Methods []Method `json:"methods,omitempty"`
}

// Value type
type Value struct {
	ID    string      `json:"id"`
	Type  string      `json:"type,omitempty"`
	Name  string      `json:"name,omitempty"`
	Unit  string      `json:"unit,omitempty"`
	Min   string      `json:"min,omitempty"`
	Max   string      `json:"max,omitempty"`
	Time  *time.Time  `json:"time,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

// Method type
type Method struct {
	Name   string  `json:"name"`
	Params []Value `json:"params,omitempty"`
}

// Param Type
type Param struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Event type: Describes an event. Priority 0 is OK, 1 is warning and 2 is handled as error
type Event struct {
	ID       string     `json:"id"`
	Message  string     `json:"message,omitempty"`
	Priority uint8      `json:"priority,omitempty"`
	Time     *time.Time `json:"time,omitempty"`
	Extra    []Param    `json:"extra,omitempty"`
}

// Meta type: Holds some meta data (maximum, minimum and average) of the sensor value
type Meta struct {
	Max float64 `json:"max,omitempty"`
	Min float64 `json:"min,omitempty"`
	Avg float64 `json:"avg,omitempty"`
	N   int     `json:"n,omitempty"`
}

// MQTTConfig type for configuration of the MQTT server
type MQTTConfig struct {
	Protocol, Server, Port, User, Password, ClientID string
}

// HomeConfig type for general configuration
type HomeConfig struct {
	DBPath   string
	MQTT     MQTTConfig
	WS       WebsocketConfig
	API      APIConfig
	Tg       TelegramConfig
	TimeZone string
	Location *time.Location
}

// WebsocketConfig type
type WebsocketConfig struct {
	Port    string
	Enabled bool
}

// APIConfig type
type APIConfig struct {
	Port    string
	Enabled bool
}

// TelegramConfig type
type TelegramConfig struct {
	Token      string
	Chats      []int64
	Enabled    bool
	RTSPSource string
}

var floatType = reflect.TypeOf(float64(0))

// GetFloat converts an interface to a float64
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

// IsNil check if the device type is nil
func (d *Device) IsNil() bool {
	if d.ID != "" ||
		d.Name != "" {
		return false
	}
	return true
}
