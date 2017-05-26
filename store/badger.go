//badger object lives in this package
package store

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"
	"reflect"
	"sync"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
	"github.com/byuoitav/monster-monitoring-service/salt"
	"github.com/dgraph-io/badger/badger"
	"github.com/dgraph-io/badger/table"
)

func Listen(events chan salt.SaltEvent, done chan bool, signal *sync.WaitGroup) {

	log.Printf("Listening for events...")

	var event salt.SaltEvent
	var listen sync.Once

	for {
		select {
		case <-done:
			listen.Do(func() {
				log.Printf("SIGTERM signal detected. Closing store...")
				Store().Close()
			})
			break
		case event = <-events:
			err := UpdateStoreBySalt(event)
			if err != nil {
				log.Printf("Error updating store: %s", err.Error())
			}
		}
	}
	signal.Done()
}
func UpdateStoreBySalt(event salt.SaltEvent) error {

	log.Printf("Adding event %s to store", event.Tag)

	return nil
}

func UpdateStoreByRoom(input base.PublicRoom) error {

	log.Printf("Updating store by room: %s in building: %s...", input.Room, input.Building)

	room := input.Building + "-" + input.Room
	key := []byte(room)

	temp, err := json.Marshal(input)
	if err != nil {
		log.Printf("Error marshaling struct to JSON: %s", err.Error())
	}

	value, _ := Store().Get(key)
	if value != nil {
		value = append(value, temp...)
	} else {
		value = temp
	}

	Store().Set(key, value)

	return nil
}

func UpdateStoreByEvent(event eventinfrastructure.Event) error {

	log.Printf("Updating store by event from device: %s...", event.Event.Device)

	room := event.Building + "-" + event.Room
	key := []byte(room)

	temp := make(map[string]string)
	temp[event.Event.EventInfoKey] = event.Event.EventInfoValue
	bytes, err := GetBytes(temp)
	if err != nil {
		return err
	}

	value, _ := Store().Get(key)
	if value != nil {
		value = append(value, bytes...)
	} else {
		value = bytes
	}

	Store().Set(key, value)

	return nil
}

//convert stuff to byte array
func GetBytes(key interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(key)
	if err != nil {
		log.Printf("Error encoding struct of type %s: %s", reflect.TypeOf(key), err.Error())
		return nil, err
	}
	return buffer.Bytes(), nil
}

//used to get instance of store
func Store() *badger.KV {
	once.Do(func() {
		store, _ = badger.NewKV(&DefaultOptions)
	})
	return store
}

//singleton instance of store
var store *badger.KV

//idiomatic way of implementing singleton pattern in golang
var once sync.Once

//should be safe, but can be tweaked
var DefaultOptions = badger.Options{
	Dir:                      "/tmp",
	DoNotCompact:             false,
	LevelOneSize:             256 << 20,
	LevelSizeMultiplier:      10,
	MapTablesTo:              table.MemoryMap,
	MaxLevels:                7,
	MaxTableSize:             64 << 20,
	MemtableSlack:            10 << 20,
	NumLevelZeroTables:       5,
	NumLevelZeroTablesStall:  10,
	NumMemtables:             5,
	SyncWrites:               false,
	ValueCompressionMinRatio: 2.0,
	ValueCompressionMinSize:  1024,
	ValueGCThreshold:         0.5,
	ValueLogFileSize:         1 << 30,
	ValueThreshold:           20,
	Verbose:                  false,
}
