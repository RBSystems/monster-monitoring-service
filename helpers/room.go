package helpers

import "github.com/byuoitav/av-api/base"

//queries the data store and returns a PublicRom
func QueryRoomStatus(building string, room string) (base.PublicRoom, error) {
	return base.PublicRoom{}, nil
}

//queries the data store and dumps room info
func GetRoomInfo(building string, room string) ([]byte, error) {
	return []byte{}, nil
}
