package storage

import (
	"time"

	"github.com/conejoninja/home/common"
)

type Storage interface {
	AddValue(id string, value common.Value) error
	AddDevice(id []byte, device common.Device) error
	GetValue(id []byte) common.Value
	GetValuesBetweenTime(id string, start, end time.Time) []common.Value
	GetDevice(id []byte) common.Device
	GetDevices() []common.Device
}
