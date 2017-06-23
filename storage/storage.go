package storage

import (
	"time"

	"github.com/conejoninja/home/common"
)

type Storage interface {
	AddValue(id string, value common.Value) error
	AddEvent(id string, value common.Event) error
	AddDevice(id []byte, device common.Device) error
	GetValue(id []byte) common.Value
	GetEvent(id []byte) common.Event
	GetValuesBetweenTime(id string, start, end time.Time) []common.Value
	GetEventsBetweenTime(id string, start, end time.Time) []common.Event
	GetDevice(id []byte) common.Device
	GetDevices() []common.Device
}
