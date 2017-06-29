package storage

import (
	"log"

	"os"

	"encoding/json"

	"strconv"
	"strings"
	"time"

	"github.com/conejoninja/home/common"
	"github.com/dgraph-io/badger/badger"
	"fmt"
)

type Badger struct {
	valuesPath  string
	devicesPath string
	metaPath    string
	eventsPath  string
	valuesKV    *badger.KV
	devicesKV   *badger.KV
	metaKV      *badger.KV
	eventsKV    *badger.KV
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

	db.eventsPath = path + "events"
	db.eventsKV = openKV(db.eventsPath)

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

func (db *Badger) AddValue(device string, value common.Value) error {

	if value.Time.IsZero() {
		now := time.Now()
		value.Time = now
	}

	id := []byte(device + "-" + value.Id + "-" + strconv.Itoa(int(value.Time.Unix())))

	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	db.valuesKV.Set(id, payload)
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

	sensor := []byte(id + "-" + strconv.Itoa(int(start.Unix())))
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
		if string(item.Key()[:len(id)])!=id || int64(timeStr) > endInt {
			nValues++
			break
		} else {
			nValues++
		}
	}

	nValuesTotal := (nValues-1)
	values := make([]common.Value, nValuesTotal)
	nValues = 0
	if nValuesTotal > 0 {
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
	}

	return values
}

func (db *Badger) AddEvent(id string, evt common.Event) error {

	var event []byte
	if evt.Time.IsZero() {
		now := time.Now()
		event = []byte(id + "-" + strconv.Itoa(int(now.Unix())))
		evt.Time = now
	} else {
		event = []byte(id + "-" + strconv.Itoa(int(evt.Time.Unix())))
	}

	payload, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	db.eventsKV.Set(event, payload)
	return nil
}

func (db *Badger) GetEvent(id []byte) common.Event {
	var evt common.Event
	itrOpt := badger.IteratorOptions{
		PrefetchSize: 1000,
		FetchValues:  true,
		Reverse:      false,
	}
	itr := db.eventsKV.NewIterator(itrOpt)
	for itr.Seek(id); itr.Valid(); itr.Next() {
		item := itr.Item()
		if string(id) == string(item.Key()) {
			err := json.Unmarshal(item.Value(), &evt)
			if err != nil {
				// Do something ?
			}
			return evt
		}
		break
	}
	return evt
}

func (db *Badger) GetMeta(id []byte) (meta common.Meta) {
	itrOpt := badger.IteratorOptions{
		PrefetchSize: 1000,
		FetchValues:  true,
		Reverse:      false,
	}
	itr := db.metaKV.NewIterator(itrOpt)
	for itr.Seek(id); itr.Valid(); itr.Next() {
		item := itr.Item()
		if string(id) == string(item.Key()) {
			err := json.Unmarshal(item.Value(), &meta)
			if err != nil {
				// Do something ?
			}
			return
		}
		break
	}
	return
}

func (db *Badger) GetEventsBetweenTime(id string, start, end time.Time) []common.Event {

	sensor := []byte(id + strconv.Itoa(int(start.Unix())))
	endInt := end.Unix()

	itrOpt := badger.IteratorOptions{
		PrefetchSize: 1000,
		FetchValues:  true,
		Reverse:      false,
	}
	itr := db.eventsKV.NewIterator(itrOpt)

	nEvents := 0
	for itr.Seek(sensor); itr.Valid(); itr.Next() {
		item := itr.Item()
		parts := strings.Split(string(item.Key()), "-")
		l := len(parts)
		timeStr, _ := strconv.Atoi(parts[l-1])
		if int64(timeStr) > endInt {
			break
		} else {
			nEvents++
		}
	}

	nEventsTotal := (nEvents-1)
	events := make([]common.Event, nEventsTotal)
	nEvents = 0
	if nEventsTotal > 0 {
		for itr.Seek(sensor); itr.Valid(); itr.Next() {
			item := itr.Item()
			err := json.Unmarshal(item.Value(), &events[nEvents])
			if err != nil {
				// Do something ?
			}
			nEvents++
			if nEvents >= nEventsTotal {
				break
			}
		}
	}

	return events
}

func (db *Badger) AddMeta(id []byte, meta common.Meta) error {
	payload, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	db.metaKV.Set(id, payload)
	return nil
}


// FOR DEBUG
func (db *Badger) ListAll(what string) {

	itrOpt := badger.IteratorOptions{
		PrefetchSize: 1000,
		FetchValues:  true,
		Reverse:      false,
	}

	var itr *badger.Iterator
	if what == "meta" {
		itr = db.metaKV.NewIterator(itrOpt)
	} else if what == "devices" {
		itr = db.devicesKV.NewIterator(itrOpt)
	} else if what == "events" {
		itr = db.eventsKV.NewIterator(itrOpt)
	} else {
		itr = db.valuesKV.NewIterator(itrOpt)
	}

	for itr.Rewind(); itr.Valid(); itr.Next() {
		item := itr.Item()
		fmt.Println(string(item.Key()), " = ", string(item.Value()))
	}

}
