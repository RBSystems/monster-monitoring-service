package accessors

type DeviceCommand struct {
	ID             int  `json:"id,omitempty"`
	DeviceID       int  `json:"device"`
	CommandID      int  `json:"command"`
	MicroserviceID int  `json:"microservice"`
	EndpointID     int  `json:"endpoint"`
	Enabled        bool `json:"enabled"`
}

func (accessorGroup *AccessorGroup) AddDeviceCommand(dc DeviceCommand) (DeviceCommand, error) {
	// devicecommand.ID needs to be changed to devicecommand.Command.ID, but Command doesn't have that field yet
	result, err := accessorGroup.Database.Exec("Insert into DeviceCommands (deviceCommandID, deviceID, commandID, microserviceID, endpointID, enabled) VALUES(?,?,?,?,?,?)", dc.ID, dc.DeviceID, dc.CommandID, dc.MicroserviceID, dc.EndpointID, dc.Enabled)

	if err != nil {
		return DeviceCommand{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return DeviceCommand{}, err
	}

	dc.ID = int(id)
	return dc, nil
}
