//badger object lives in this package
package badger

import (
	"log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
	"github.com/dgraph-io/badger/badger"
)

var store badger.KV

func Init() {
}

func UpdateStoreByRoom(room base.PublicRoom) error {

	log.Printf("Updating store by room: %s in building: %s...", room.Room, room.Building)

	return nil
}

func UpdateStoreByEvent(event eventinfrastructure.Event) error {

	log.Printf("Updating store by event from device: %s...", event.Event.Device)
	return nil
}

func Listen() {
}
