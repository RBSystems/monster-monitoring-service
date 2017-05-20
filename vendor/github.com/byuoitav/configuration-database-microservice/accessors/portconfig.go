package accessors

import (
	"database/sql"
)

type PortConfiguration struct {
	ID                  int `json:"id,omitempty"`
	DestinationDeviceID int `json:"destination-device"`
	PortID              int `json:"port"`
	SourceDeviceID      int `json:"source-device"`
	HostDeviceID        int `json:"host-device"`
}

func (accessorGroup *AccessorGroup) GetPortConfiguration(building string, room string, device string) ([]PortConfiguration, error) {
	rows, err := accessorGroup.Database.Query("SELECT * FROM PortConfiguration")
	if err != nil {
		return []PortConfiguration{}, err
	}

	portconfigurations, err := exctractPortConfigurationData(rows)
	if err != nil {
		return []PortConfiguration{}, err
	}
	defer rows.Close()

	return portconfigurations, nil
}

func (accessorGroup *AccessorGroup) AddPortConfiguration(pc PortConfiguration) (PortConfiguration, error) {
	response, err := accessorGroup.Database.Exec("INSERT INTO PortConfiguration (portConfigurationID, destinationDeviceID, portID, sourceDeviceID, hostDeviceID) VALUES(?,?,?,?,?)", pc.ID, pc.DestinationDeviceID, pc.PortID, pc.SourceDeviceID, pc.HostDeviceID)
	if err != nil {
		return PortConfiguration{}, err
	}

	id, err := response.LastInsertId()
	pc.ID = int(id)

	return pc, nil
}

func exctractPortConfigurationData(rows *sql.Rows) ([]PortConfiguration, error) {
	var portconfigurations []PortConfiguration
	var portconfiguration PortConfiguration
	var id *int
	var ddID *int
	var pID *int
	var sdID *int
	var hID *int

	for rows.Next() {
		err := rows.Scan(&id, &ddID, &pID, &sdID, &hID)
		if err != nil {
			return []PortConfiguration{}, err
		}

		if id != nil {
			portconfiguration.ID = *id
		}
		if ddID != nil {
			portconfiguration.DestinationDeviceID = *ddID
		}
		if pID != nil {
			portconfiguration.PortID = *pID
		}
		if sdID != nil {
			portconfiguration.SourceDeviceID = *sdID
		}
		if hID != nil {
			portconfiguration.HostDeviceID = *hID
		}

		portconfigurations = append(portconfigurations, portconfiguration)
	}

	err := rows.Err()
	if err != nil {
		return []PortConfiguration{}, err
	}

	return portconfigurations, nil
}
