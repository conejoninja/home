package storage

import (
	"log"

	"os"

	"encoding/json"

	"github.com/conejoninja/home/common"
	"github.com/dgraph-io/badger/badger"
	"time"
	"strconv"
	"strings"
)

type Badger struct {
	valuesPath  string
	devicesPath string
	metaPath    string
	valuesKV    *badger.KV
	devicesKV   *badger.KV
	metaKV      *badger.KV
}

func NewBadger(path string) *Badger {

	l := len(path)
	if string(path[l-1]) != "/" {
		path += "/"
	}

	var db Badger
	db.valuesPath = path + "values"
	db.valuesKV = openKV(db.valuesPath)

	db.devicesPath = path + "devices"
	db.devicesKV = openKV(db.devicesPath)

	db.metaPath = path + "meta"
	db.metaKV = openKV(db.metaPath)

	return &db
}

func openKV(path string) *badger.KV {
	err := os.MkdirAll(path, 0777)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	opt := badger.DefaultOptions
	opt.Dir = path
	opt.SyncWrites = true
	kv, err := badger.NewKV(&opt)
	if err != nil {
		log.Fatal(err)
	}
	return kv
}

func (db *Badger) Close() {
	db.Close()
}

func (db *Badger) AddDevice(id []byte, device common.Device) error {
	payload, err := json.Marshal(device)
	if err != nil {
		return err
	}
	db.devicesKV.Set(id, payload)
	return nil
}

func (db *Badger) GetDevice(id []byte) common.Device {
	var device common.Device
	itrOpt := badger.IteratorOptions{
		PrefetchSize: 1000,
		FetchValues:  true,
		Reverse:      false,
	}
	itr := db.devicesKV.NewIterator(itrOpt)
	for itr.Seek(id); itr.Valid(); itr.Next() {
		item := itr.Item()
		if string(id) == string(item.Key()) {
			err := json.Unmarshal(item.Value(), &device)
			if err != nil {
				// Do something ?
			}
			return device
		}
		break
	}
	return device
}

func (db *Badger) GetDevices() []common.Device {
	itrOpt := badger.IteratorOptions{
		PrefetchSize: 1000,
		FetchValues:  true,
		Reverse:      false,
	}
	itr := db.devicesKV.NewIterator(itrOpt)
	nDevices := 0
	for itr.Rewind(); itr.Valid(); itr.Next() {
		nDevices++
	}
	devices := make([]common.Device, nDevices)
	nDevices = 0
	for itr.Rewind(); itr.Valid(); itr.Next() {
		item := itr.Item()
		err := json.Unmarshal(item.Value(), &devices[nDevices])
		if err != nil {
			// Do something ?
		}
		nDevices++
	}
	return devices
}


func (db *Badger) AddValue(id string, value common.Value) error {
	now := time.Now()
	sensor := []byte(id + "-" + strconv.Itoa(int(now.Unix())))

	value.Time = now

	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	db.valuesKV.Set(sensor, payload)
	return nil
}


func (db *Badger) GetValue(id []byte) common.Value {
	var value common.Value
	itrOpt := badger.IteratorOptions{
		PrefetchSize: 1000,
		FetchValues:  true,
		Reverse:      false,
	}
	itr := db.valuesKV.NewIterator(itrOpt)
	for itr.Seek(id); itr.Valid(); itr.Next() {
		item := itr.Item()
		if string(id) == string(item.Key()) {
			err := json.Unmarshal(item.Value(), &value)
			if err != nil {
				// Do something ?
			}
			return value
		}
		break
	}
	return value
}


func (db *Badger) GetValuesBetweenTime(id string, start, end time.Time) []common.Value {

	sensor := []byte(id + strconv.Itoa(int(start.Unix())))
	endInt := end.Unix()

	itrOpt := badger.IteratorOptions{
		PrefetchSize: 1000,
		FetchValues:  true,
		Reverse:      false,
	}
	itr := db.valuesKV.NewIterator(itrOpt)

	nValues := 0
	for itr.Seek(sensor); itr.Valid(); itr.Next() {
		item := itr.Item()
		parts := strings.Split(string(item.Key()), "-")
		l := len(parts)
		timeStr, _ := strconv.Atoi(parts[l-1])
		if int64(timeStr) > endInt {
			break
		} else {
			nValues++
		}
	}

	nValuesTotal := nValues
	values := make([]common.Value, nValuesTotal)
	nValues = 0
	for itr.Seek(sensor); itr.Valid(); itr.Next() {
		item := itr.Item()
		err := json.Unmarshal(item.Value(), &values[nValues])
		if err != nil {
			// Do something ?
		}
		nValues++
		if nValues >= nValuesTotal {
			break
		}
	}

	return values
}
