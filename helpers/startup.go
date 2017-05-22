package helpers

import (
	"log"

	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/av-api/status"
	"github.com/byuoitav/monster-monitoring-service/badger"
)

func OnStart() {

	log.Printf("Querying buildings...")

	buildings, err := dbo.GetBuildings()
	if err != nil {
		log.Printf("Error getting buildings from database: %s", err.Error())
	}

	for _, building := range buildings {

		log.Printf("Getting rooms from building %s...", building.Name)
		rooms, err := dbo.GetRoomsByBuilding(building.Name)
		if err != nil {
			log.Printf("Error getting rooms from %s: %s", building.Name, err.Error())
		}

		for _, room := range rooms {

			log.Printf("Getting status of room %s...", room.Name)
			roomStatus, err := status.GetRoomStatus(building.Name, room.Name)
			if err != nil {
				log.Printf("Error getting status for room: %s in building %s: %s", building.Name, room.Name, err.Error())
			}

			log.Printf("Adding room %s to Badger...", room.Name)
			err = badger.UpdateStoreByRoom(roomStatus)
			if err != nil {
				log.Printf("Error adding room: %s in building: %s to Badger: %s", building.Name, room.Name, err.Error())
			}

		}

	}

	return
}
