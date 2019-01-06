package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/conejoninja/home/common"
	"github.com/dgraph-io/badger"
)

//Badger type
type Badger struct {
	valuesPath  string
	devicesPath string
	metaPath    string
	eventsPath  string
	valuesDB    *badger.DB
	devicesDB   *badger.DB
	metaDB      *badger.DB
	eventsDB    *badger.DB
}

// NewBadger opens and returns a storage
func NewBadger(path string) *Badger {

	l := len(path)
	if string(path[l-1]) != "/" {
		path += "/"
	}

	var db Badger
	db.valuesPath = path + "values"
	db.valuesDB = openDB(db.valuesPath)

	db.devicesPath = path + "devices"
	db.devicesDB = openDB(db.devicesPath)

	db.metaPath = path + "meta"
	db.metaDB = openDB(db.metaPath)

	db.eventsPath = path + "events"
	db.eventsDB = openDB(db.eventsPath)

	return &db
}

func openDB(path string) *badger.DB {
	err := os.MkdirAll(path+"/", 0777)
	if err != nil {
		log.Fatal(err)
	}

	opt := badger.DefaultOptions
	opt.Dir = path
	opt.ValueDir = path
	opt.SyncWrites = true
	db, err := badger.Open(opt)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// Close the storage
func (db *Badger) Close() {
	db.Close()
}

// AddDevice adds a new device
func (db *Badger) AddDevice(id []byte, device common.Device) error {
	payload, err := json.Marshal(device)
	if err != nil {
		return err
	}
	txn := db.devicesDB.NewTransaction(true)
	defer txn.Discard()
	txn.Set(id, payload)
	txn.Commit()
	return nil
}

// GetDevice returns a device given its ID
func (db *Badger) GetDevice(id []byte) common.Device {
	var device common.Device
	itrOpt := badger.IteratorOptions{
		PrefetchSize: 1000,
		Reverse:      false,
	}

	txn := db.devicesDB.NewTransaction(true)
	defer txn.Discard()
	itr := txn.NewIterator(itrOpt)
	for itr.Seek(id); itr.Valid(); itr.Next() {
		item := itr.Item()
		if string(id) == string(item.Key()) {
			var value []byte
			err := item.Value(func(v []byte) error {
				value = v
				return nil
			})
			if err != nil {
				return device
			}
			err = json.Unmarshal(value, &device)
			if err != nil {
				// Do something ?
			}
			return device
		}
		break
	}
	return device
}

// GetDevices returns all the devices in the network
func (db *Badger) GetDevices() []common.Device {
	itrOpt := badger.IteratorOptions{
		PrefetchSize: 1000,
		Reverse:      false,
	}
	txn := db.devicesDB.NewTransaction(true)
	defer txn.Discard()
	itr := txn.NewIterator(itrOpt)
	nDevices := 0
	for itr.Rewind(); itr.Valid(); itr.Next() {
		nDevices++
	}
	devices := make([]common.Device, nDevices)
	nDevices = 0
	for itr.Rewind(); itr.Valid(); itr.Next() {
		item := itr.Item()
		var value []byte
		err := item.Value(func(v []byte) error {
			value = v
			return nil
		})
		if err != nil {
			return devices
		}
		err = json.Unmarshal(value, &devices[nDevices])
		if err != nil {
			// Do something ?
		}
		nDevices++
	}
	return devices
}

// AddValue adds a sensor value to the storage
func (db *Badger) AddValue(device string, value common.Value) error {

	if value.Time == nil || (*value.Time).IsZero() {
		now := time.Now()
		value.Time = &now
	}

	id := []byte(device + "-" + value.ID + "-" + strconv.Itoa(int(value.Time.Unix())))

	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	txn := db.valuesDB.NewTransaction(true)
	defer txn.Discard()
	txn.Set(id, payload)
	txn.Commit()
	return nil
}

// GetValue returns a specific sensor value
func (db *Badger) GetValue(id []byte) common.Value {
	var value common.Value
	itrOpt := badger.IteratorOptions{
		PrefetchSize: 1000,
		Reverse:      false,
	}
	txn := db.valuesDB.NewTransaction(true)
	defer txn.Discard()
	itr := txn.NewIterator(itrOpt)
	for itr.Seek(id); itr.Valid(); itr.Next() {
		item := itr.Item()
		var itemValue []byte
		err := item.Value(func(v []byte) error {
			itemValue = v
			return nil
		})
		if err != nil {
			return value
		}
		if string(id) == string(item.Key()) {
			err := json.Unmarshal(itemValue, &value)
			if err != nil {
				// Do something ?
			}
			return value
		}
		break
	}
	return value
}

// GetLastValue returns the last value of a sensor given its ID
func (db *Badger) GetLastValue(id string) common.Value {
	var value common.Value
	itrOpt := badger.IteratorOptions{
		PrefetchSize: 1000,
		Reverse:      true,
	}
	txn := db.valuesDB.NewTransaction(true)
	defer txn.Discard()
	itr := txn.NewIterator(itrOpt)
	for itr.Seek([]byte(id + "-9")); itr.Valid(); itr.Next() {
		item := itr.Item()
		var itemValue []byte
		err := item.Value(func(v []byte) error {
			itemValue = v
			return nil
		})
		if err != nil {
			return value
		}
		if id == string(item.Key()[:len(id)]) {
			err := json.Unmarshal(itemValue, &value)
			if err != nil {
				// Do something ?
			}
			return value
		}
		break
	}
	return value
}

// GetValuesBetweenTime returns all the values between two given dates
func (db *Badger) GetValuesBetweenTime(id string, start, end time.Time) []common.Value {

	sensor := []byte(id + "-" + strconv.Itoa(int(start.Unix())))
	endInt := end.Unix()

	itrOpt := badger.IteratorOptions{
		PrefetchSize: 1000,
		Reverse:      false,
	}
	txn := db.valuesDB.NewTransaction(true)
	defer txn.Discard()
	itr := txn.NewIterator(itrOpt)
	nValues := 0
	for itr.Seek(sensor); itr.Valid(); itr.Next() {
		item := itr.Item()
		parts := strings.Split(string(item.Key()), "-")
		l := len(parts)
		timeStr, _ := strconv.Atoi(parts[l-1])
		if string(item.Key()[:len(id)]) != id || int64(timeStr) > endInt {
			nValues++
			break
		} else {
			nValues++
		}
	}

	nValuesTotal := nValues - 1
	values := make([]common.Value, nValuesTotal)
	nValues = 0
	if nValuesTotal > 0 {
		for itr.Seek(sensor); itr.Valid(); itr.Next() {
			item := itr.Item()
			var value []byte
			err := item.Value(func(v []byte) error {
				value = v
				return nil
			})
			if err != nil {
				return values
			}
			err = json.Unmarshal(value, &values[nValues])
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

// AddEvent adds an event
func (db *Badger) AddEvent(id string, evt common.Event) error {

	var event []byte
	if evt.Time == nil || (*evt.Time).IsZero() {
		now := time.Now()
		event = []byte(id + "-" + strconv.Itoa(int(now.Unix())))
		evt.Time = &now
	} else {
		event = []byte(id + "-" + strconv.Itoa(int(evt.Time.Unix())))
	}

	payload, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	txn := db.eventsDB.NewTransaction(true)
	defer txn.Discard()
	txn.Set(event, payload)
	txn.Commit()
	return nil
}

// GetEvent returns a specific event
func (db *Badger) GetEvent(id []byte) common.Event {
	var evt common.Event
	itrOpt := badger.IteratorOptions{
		PrefetchSize: 1000,
		Reverse:      false,
	}
	txn := db.eventsDB.NewTransaction(true)
	defer txn.Discard()
	itr := txn.NewIterator(itrOpt)
	for itr.Seek(id); itr.Valid(); itr.Next() {
		item := itr.Item()
		var value []byte
		err := item.Value(func(v []byte) error {
			value = v
			return nil
		})
		if err != nil {
			return evt
		}
		if string(id) == string(item.Key()) {
			err := json.Unmarshal(value, &evt)
			if err != nil {
				// Do something ?
			}
			return evt
		}
		break
	}
	return evt
}

// GetLastEvents returns a given number of most recent events
func (db *Badger) GetLastEvents(id string, count int) []common.Event {
	evts := make([]common.Event, count)
	itrOpt := badger.IteratorOptions{
		PrefetchSize: 1000,
		Reverse:      true,
	}
	txn := db.eventsDB.NewTransaction(true)
	defer txn.Discard()
	itr := txn.NewIterator(itrOpt)
	e := 0
	for itr.Seek([]byte(id + "-9")); itr.Valid(); itr.Next() {
		item := itr.Item()
		var value []byte
		err := item.Value(func(v []byte) error {
			value = v
			return nil
		})
		if err != nil {
			return evts
		}
		if string(id) == string(item.Key()[:len(id)]) {
			err := json.Unmarshal(value, &evts[e])
			if err != nil {
				// Do something ?
			}
		} else {
			break
		}
		e++
		if e >= count {
			break
		}
	}
	return evts[:e]
}

// GetMeta returns a specific Meta type (max., min., avg.) of a sensor
func (db *Badger) GetMeta(id []byte) (meta common.Meta) {
	itrOpt := badger.IteratorOptions{
		PrefetchSize: 1000,
		Reverse:      false,
	}
	txn := db.metaDB.NewTransaction(true)
	defer txn.Discard()
	itr := txn.NewIterator(itrOpt)
	for itr.Seek(id); itr.Valid(); itr.Next() {
		item := itr.Item()
		var value []byte
		err := item.Value(func(v []byte) error {
			value = v
			return nil
		})
		if err != nil {
			return
		}
		if string(id) == string(item.Key()) {
			err := json.Unmarshal(value, &meta)
			if err != nil {
				// Do something ?
			}
			return
		}
		break
	}
	return
}

// GetEventsBetweenTime returns all the events between two given dates
func (db *Badger) GetEventsBetweenTime(id string, start, end time.Time) []common.Event {

	sensor := []byte(id + strconv.Itoa(int(start.Unix())))
	endInt := end.Unix()

	itrOpt := badger.IteratorOptions{
		PrefetchSize: 1000,
		Reverse:      false,
	}
	txn := db.eventsDB.NewTransaction(true)
	defer txn.Discard()
	itr := txn.NewIterator(itrOpt)
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

	nEventsTotal := (nEvents - 1)
	events := make([]common.Event, nEventsTotal)
	nEvents = 0
	if nEventsTotal > 0 {
		for itr.Seek(sensor); itr.Valid(); itr.Next() {
			item := itr.Item()
			var value []byte
			err := item.Value(func(v []byte) error {
				value = v
				return nil
			})
			if err != nil {
				return events
			}
			err = json.Unmarshal(value, &events[nEvents])
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

// AddMeta adds a Meta type to the storage
func (db *Badger) AddMeta(id []byte, meta common.Meta) error {
	payload, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	txn := db.metaDB.NewTransaction(true)
	defer txn.Discard()
	txn.Set(id, payload)
	txn.Commit()
	return nil
}

// ListAll lists all the pairs DB of a given type
func (db *Badger) ListAll(what string) {
	itrOpt := badger.IteratorOptions{
		PrefetchSize: 1000,
		Reverse:      false,
	}

	var txn *badger.Txn
	if what == "meta" {
		txn = db.metaDB.NewTransaction(true)
	} else if what == "devices" {
		txn = db.devicesDB.NewTransaction(true)
	} else if what == "events" {
		txn = db.eventsDB.NewTransaction(true)
	} else {
		txn = db.valuesDB.NewTransaction(true)
	}
	defer txn.Discard()
	itr := txn.NewIterator(itrOpt)
	for itr.Rewind(); itr.Valid(); itr.Next() {
		item := itr.Item()
		var value []byte
		err := item.Value(func(v []byte) error {
			value = v
			return nil
		})
		if err != nil {
			return
		}
		fmt.Println(string(item.Key()), " = ", string(value))
	}
}
