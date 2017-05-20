package accessors

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

//Device represents a device object as found in the DB.
type Device struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name,omitempty"`
	Address     string    `json:"address"`
	Input       bool      `json:"input"`
	Output      bool      `json:"output"`
	Building    Building  `json:"building"`
	Room        Room      `json:"room"`
	Type        string    `json:"type"`
	Power       string    `json:"power"`
	Roles       []string  `json:"roles,omitempty"`
	Blanked     *bool     `json:"blanked,omitempty"`
	Volume      *int      `json:"volume,omitempty"`
	Muted       *bool     `json:"muted,omitempty"`
	PowerStates []string  `json:"powerstates,omitempty"`
	Responding  bool      `json:"responding"`
	Ports       []Port    `json:"ports,omitempty"`
	Commands    []Command `json:"commands,omitempty"`
}

//GetFullName reutrns the string of building + room + name
func (d *Device) GetFullName() string {
	return (d.Building.Shortname + "-" + d.Room.Name + "-" + d.Name)
}

//Port represents a physical port on a device (HDMI, DP, Audo, etc.)
//TODO: this corresponds to the PortConfiguration table in the database!!!
type Port struct {
	Source      string `json:"source"`
	Name        string `json:"name"`
	Destination string `json:"destination"`
	Host        string `json:"host"`
}

//Endpoint represents a path on a microservice.
type Endpoint struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

/*
GetDevicesByQuery is a function that abstracts some of the execution and extraction
of data from the database when we're looking for responses based on the COMPLETE device struct.
The function MAY have the WHERE clause passed in to limit the devices found.
The function MAY have any JOIN clauses necessary to the WEHRE Clause not included in
the base query.
JOIN statements in the base query:
JOIN Rooms on Devices.roomID = Rooms.RoomID
JOIN Buildings on Rooms.buildingID = Buildings.buildingID
JOIN DeviceTypes on Devices.typeID = DeviceTypes.deviceTypeID
If empty string is passed in no WHERE clause will be appended, and thus all devices
will be returned.

Flow	->	Find all devices based on the clause passed in
			->	For each device found find the Ports
			->	For each device found find the Commands

Examples of valid parameters.
Example 1:
`JOIN deviceRole on deviceRole.deviceID = Devices.deviceID
JOIN DeviceRoleDefinition on DeviceRole.deviceRoleDefinitionID = DeviceRoleDefinition.deviceRoleDefinitionID
WHERE DeviceRoleDefinition.name LIKE 'AudioIn'`
Example 2:
`WHERE Devices.RoomID = 1`
*/
func (accessorGroup *AccessorGroup) GetDevicesByQuery(query string, parameters ...interface{}) ([]Device, error) {
	baseQuery := `SELECT DISTINCT Devices.deviceID,
  	Devices.Name as deviceName,
  	Devices.address as deviceAddress,
  	Devices.input,
  	Devices.output,
	Devices.displayName,
  	Rooms.roomID,
  	Rooms.name as roomName,
  	Rooms.description as roomDescription,
  	Buildings.buildingID,
  	Buildings.name as buildingName,
  	Buildings.shortName as buildingShortname,
  	Buildings.description as buildingDescription,
  	DeviceTypes.name as deviceType
  	FROM Devices
  	JOIN Rooms on Rooms.roomID = Devices.roomID
  	JOIN Buildings on Buildings.buildingID = Devices.buildingID
  	JOIN DeviceTypes on Devices.typeID = DeviceTypes.deviceTypeID
    JOIN DeviceRole on DeviceRole.deviceID = Devices.deviceID
    JOIN DeviceRoleDefinition on DeviceRole.deviceRoleDefinitionID = DeviceRoleDefinition.deviceRoleDefinitionID`

	allDevices := []Device{}

	rows, err := accessorGroup.Database.Query(baseQuery+" "+query, parameters...)
	if err != nil {
		return []Device{}, err
	}

	defer rows.Close()

	for rows.Next() {

		device := Device{}

		err := rows.Scan(&device.ID,
			&device.Name,
			&device.Address,
			&device.Input,
			&device.Output,
			&device.DisplayName,
			&device.Room.ID,
			&device.Room.Name,
			&device.Room.Description,
			&device.Building.ID,
			&device.Building.Name,
			&device.Building.Shortname,
			&device.Building.Description,
			&device.Type)
		if err != nil {
			return []Device{}, err
		}

		device.Commands, err = accessorGroup.GetDeviceCommandsByBuildingAndRoomAndName(device.Building.Shortname, device.Room.Name, device.Name)
		if err != nil {
			return []Device{}, err
		}

		device.Ports, err = accessorGroup.GetDevicePortsByBuildingAndRoomAndName(device.Building.Shortname, device.Room.Name, device.Name)
		if err != nil {
			return []Device{}, err
		}

		device.PowerStates, err = accessorGroup.GetPowerStatesByDeviceID(device.ID)
		if err != nil {
			return []Device{}, err
		}

		device.Roles, err = accessorGroup.GetRolesByDeviceID(device.ID)
		if err != nil {
			return []Device{}, err
		}

		allDevices = append(allDevices, device)
	}

	return allDevices, nil
}

func (AccessorGroup *AccessorGroup) GetRolesByDeviceID(deviceID int) ([]string, error) {
	query := `Select DeviceRoleDefinition.Name From DeviceRoleDefinition 
	JOIN DeviceRole dr on dr.deviceRoleDefinitionID = DeviceRoleDefinition.deviceRoleDefinitionID 
	WHERE dr.deviceID = ?`

	toReturn := []string{}

	rows, err := AccessorGroup.Database.Query(query, deviceID)
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var value string

		err = rows.Scan(&value)
		if err != nil {
			return []string{}, err
		}
		toReturn = append(toReturn, value)
	}
	return toReturn, nil
}

//GetPowerStatesByDeviceID gets the powerstates allowed for a given devices based on the
//DevicePowerStates table in the DB.
func (AccessorGroup *AccessorGroup) GetPowerStatesByDeviceID(deviceID int) ([]string, error) {
	query := `SELECT PowerStates.name FROM PowerStates
	JOIN DevicePowerStates on DevicePowerStates.powerStateID = PowerStates.powerStateID
	Where DevicePowerStates.deviceID = ?`

	toReturn := []string{}
	rows, err := AccessorGroup.Database.Query(query, deviceID)
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var value string

		err := rows.Scan(&value)
		if err != nil {
			return []string{}, err
		}
		toReturn = append(toReturn, value)
	}
	return toReturn, nil
}

//GetDevicesByBuildingAndRoomAndRole gets the devices in the room specified with the given role,
//as specified in the DeviceRole table in the DB
func (accessorGroup *AccessorGroup) GetDevicesByBuildingAndRoomAndRole(buildingShortname string, roomName string, roleName string) ([]Device, error) {
	log.Printf("Getting ")
	devices, err := accessorGroup.GetDevicesByQuery(`WHERE Rooms.name LIKE ? AND Buildings.shortname LIKE ? AND DeviceRoleDefinition.name LIKE ?`,
		roomName, buildingShortname, roleName)

	if err != nil {
		log.Printf("Error: %v", err.Error())
		return []Device{}, err
	}
	switch strings.ToLower(roleName) {

	}
	return devices, nil

}

//GetDevicesByRoleAndType Gets all teh devices that have a given role and type.
func (accessorGroup *AccessorGroup) GetDevicesByRoleAndType(deviceRole string, deviceType string, production string) ([]Device, error) {
	return accessorGroup.GetDevicesByQuery(`WHERE DeviceRoleDefinition.name LIKE ? AND DeviceTypes.name LIKE ? AND Rooms.roomDesignation = ?`, deviceRole, deviceType, production)
}

//GetDevicesByBuildingAndRoom get all the devices in the room specified.
func (accessorGroup *AccessorGroup) GetDevicesByBuildingAndRoom(buildingShortname string, roomName string) ([]Device, error) {
	log.Printf("Getting devices in room %s and building %s", roomName, buildingShortname)

	devices, err := accessorGroup.GetDevicesByQuery(
		`WHERE Rooms.name=? AND Buildings.shortName=?`, roomName, buildingShortname)

	if err != nil {
		return []Device{}, err
	}

	return devices, nil
}

//GetDeviceCommandsByBuildingAndRoomAndName gets all the commands for the device
//specified. Note that we assume that device names are unique within a room.
func (accessorGroup *AccessorGroup) GetDeviceCommandsByBuildingAndRoomAndName(buildingShortname string, roomName string, deviceName string) ([]Command, error) {
	allCommands := []Command{}
	rows, err := accessorGroup.Database.Query(`SELECT Commands.name as commandName, Endpoints.name as endpointName, Endpoints.path as endpointPath, Microservices.address as microserviceAddress
    FROM Devices
    JOIN DeviceCommands on Devices.deviceID = DeviceCommands.deviceID JOIN Commands on DeviceCommands.commandID = Commands.commandID JOIN Endpoints on DeviceCommands.endpointID = Endpoints.endpointID JOIN Microservices ON DeviceCommands.microserviceID = Microservices.microserviceID
    JOIN Rooms ON Rooms.roomID=Devices.roomID
    JOIN Buildings ON Rooms.buildingID=Buildings.buildingID
    WHERE Rooms.name=? AND Buildings.shortName=? AND Devices.name=?`, roomName, buildingShortname, deviceName)
	if err != nil {
		return []Command{}, err
	}
	defer rows.Close()

	allCommands, err = ExtractCommand(rows)
	if err != nil {
		return allCommands, err
	}

	return allCommands, nil
}

//GetDevicePortsByBuildingAndRoomAndName gets the ports for the device
//specified. Note that we assume that device names are unique within a room.
func (accessorGroup *AccessorGroup) GetDevicePortsByBuildingAndRoomAndName(buildingShortname string, roomName string, deviceName string) ([]Port, error) {
	allPorts := []Port{}

	rows, err := accessorGroup.Database.Query(`SELECT srcDevice.Name as sourceName, Ports.name as portName, destDevice.Name as DestinationDevice, hostDevice.name as HostDevice FROM Ports
    JOIN PortConfiguration ON Ports.PortID = PortConfiguration.PortID
    JOIN Devices as srcDevice on srcDevice.DeviceID = PortConfiguration.sourceDeviceID
    JOIN Devices as destDevice on destDevice.DeviceID = PortConfiguration.destinationDeviceID
		JOIN Devices as hostDevice on hostDevice.DeviceID = PortConfiguration.hostDeviceID
    JOIN Rooms ON Rooms.roomID=destDevice.roomID
    JOIN Buildings ON Rooms.buildingID=Buildings.buildingID
    WHERE Rooms.name=? AND Buildings.shortName=? AND hostDevice.name=?`, roomName, buildingShortname, deviceName)
	if err != nil {
		log.Print(err)
		return []Port{}, err
	}
	defer rows.Close()

	for rows.Next() {
		port := Port{}

		err := rows.Scan(&port.Source, &port.Name, &port.Destination, &port.Host)
		if err != nil {
			log.Print(err)
			return []Port{}, err
		}

		allPorts = append(allPorts, port)
	}

	return allPorts, nil
}

//GetDeviceByBuildingAndRoomAndName gets the device
//specified. Note that we assume that device names are unique within a room.
func (accessorGroup *AccessorGroup) GetDeviceByBuildingAndRoomAndName(buildingShortname string, roomName string, deviceName string) (Device, error) {
	dev, err := accessorGroup.GetDevicesByQuery("WHERE Buildings.shortName = ? AND Rooms.name = ? AND Devices.name = ?", buildingShortname, roomName, deviceName)
	if err != nil || len(dev) == 0 {
		return Device{}, err
	}

	return dev[0], nil
}

//PutDeviceAttributeByDeviceAndRoomAndBuilding allows you to change attribute values for devices
//Currently sets volume and muted.
func (accessorGroup *AccessorGroup) PutDeviceAttributeByDeviceAndRoomAndBuilding(building string, room string, device string, attribute string, attributeValue string) (Device, error) {
	switch strings.ToLower(attribute) {
	case "volume":
		statement := `update AudioDevices SET volume = ? WHERE deviceID =
			(Select deviceID from Devices
				JOIN Rooms on Rooms.roomID = Devices.roomID
				JOIN Buildings on Buildings.buildingID = Rooms.buildingID
				WHERE Devices.name LIKE ? AND Rooms.name LIKE ? AND Buildings.shortName LIKE ?)`
		val, err := strconv.Atoi(attributeValue)
		if err != nil {
			return Device{}, err
		}

		_, err = accessorGroup.Database.Exec(statement, val, device, room, building)
		if err != nil {
			return Device{}, err
		}
		break

	case "muted":
		var valToSet bool
		switch attributeValue {
		case "true":
			valToSet = true
			break
		case "false":
			valToSet = false
			break
		default:
			return Device{}, errors.New("Invalid attribute value, must be a boolean.")
		}
		statement := `update AudioDevices SET muted = ? WHERE deviceID =
			(Select deviceID from Devices
				JOIN Rooms on Rooms.roomID = Devices.roomID
				JOIN Buildings on Buildings.buildingID = Rooms.buildingID
				WHERE Devices.name LIKE ? AND Rooms.name LIKE ? AND Buildings.shortName LIKE ?)`
		_, err := accessorGroup.Database.Exec(statement, valToSet, device, room, building)
		if err != nil {
			return Device{}, err
		}
		break
	}

	dev, err := accessorGroup.GetDeviceByBuildingAndRoomAndName(building, room, device)
	return dev, err
}

func (accessorGroup *AccessorGroup) AddDevice(d Device) (Device, error) {
	log.Printf("Adding device %v to room %v in building %v", d.Name, d.Room.Name, d.Building.Shortname)

	// get device type string, put it into d.Type
	dt, err := accessorGroup.GetDeviceTypeByName(d.Type)
	if err != nil {
		return Device{}, err
	}

	// if device already exists in database, stop
	_, err = accessorGroup.GetDeviceByBuildingAndRoomAndName(d.Building.Shortname, d.Room.Name, d.Name)
	if err != nil {
		return Device{}, fmt.Errorf("device already exists in room, please choose a different name")
	}

	// insert into devices
	result, err := accessorGroup.Database.Exec("Insert into Devices (name, address, input, output, buildingID, roomID, typeID) VALUES (?,?,?,?,?,?,?)", d.Name, d.Address, d.Input, d.Output, d.Building.ID, d.Room.ID, dt.ID)
	if err != nil {
		return Device{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Device{}, err
	}

	d.ID = int(id)

	// insert the roles into the DeviceRole table
	var deviceroles []DeviceRole
	for _, role := range d.Roles {
		r, err := accessorGroup.GetDeviceRoleDefByName(role)
		if err != nil {
			return Device{}, fmt.Errorf("device role definition: %v does not exist", role)
		}
		var dr DeviceRole
		dr.DeviceID = d.ID
		dr.DeviceRoleDefinitionID = r.ID

		deviceroles = append(deviceroles, dr)
	}

	// insert the powerstates into the DevicePowerStates table
	var devicepowerstates []DevicePowerState
	for _, ps := range d.PowerStates {
		p, err := accessorGroup.GetPowerStateByName(ps)
		if err != nil {
			return Device{}, fmt.Errorf("powerstate: %v does not exist", ps)
		}
		var dps DevicePowerState
		dps.DeviceID = d.ID
		dps.PowerStateID = p.ID

		devicepowerstates = append(devicepowerstates, dps)
	}

	// insert the ports into the PortConfiguration table
	var portconfigurations []PortConfiguration
	for _, port := range d.Ports {
		// get portID
		pt, err := accessorGroup.GetPortTypeByName(port.Name)
		if err != nil {
			return Device{}, fmt.Errorf("port type: %v does not exist", port.Name)
		}

		// get sourceDeviceID
		sd, err := accessorGroup.GetDeviceByBuildingAndRoomAndName(d.Building.Shortname, d.Room.Name, port.Source)
		if err != nil {
			return Device{}, fmt.Errorf("source device %v does not exist in this room", port.Source)
		}

		// get destinationDeviceID
		dd, err := accessorGroup.GetDeviceByBuildingAndRoomAndName(d.Building.Shortname, d.Room.Name, port.Destination)
		if err != nil {
			return Device{}, fmt.Errorf("destination device %v does not exist in this room", port.Destination)
		}

		// get hostDeviceID
		//		hd, err := accessorGroup.GetDeviceByBuildingAndRoomAndName(d.Building.Shortname, d.Room.Name, port.Host)
		//		if err != nil {
		//			return Device{}, fmt.Errorf("host device %v does not exist in this room", port.Host)
		//		}

		var p PortConfiguration
		p.PortID = pt.ID
		p.SourceDeviceID = sd.ID
		p.DestinationDeviceID = dd.ID
		//		p.HostDeviceID = hd.ID
		p.HostDeviceID = d.ID // always the current device you are adding?

		portconfigurations = append(portconfigurations, p)
	}

	// insert the comamnds into the DeviceCommands table
	var devicecommands []DeviceCommand
	for index, command := range d.Commands {
		// get commandID
		rc, err := accessorGroup.GetRawCommandByName(command.Name)
		if err != nil {
			return Device{}, fmt.Errorf("raw command: %v does not exist", command.Name)
		}

		// get endpoint
		ep, err := accessorGroup.GetEndpointByName(command.Endpoint.Name)
		if err != nil {
			return Device{}, fmt.Errorf("endpoint: %v does not exist", command.Endpoint.Name)
		}

		// get microserviceID
		mc, err := accessorGroup.GetMicroserviceByAddress(command.Microservice)
		if err != nil {
			return Device{}, fmt.Errorf("microservice address: %v does not exist", command.Microservice)
		}

		var dc DeviceCommand
		dc.DeviceID = d.ID
		dc.CommandID = rc.ID
		dc.MicroserviceID = mc.ID
		dc.EndpointID = ep.ID
		dc.Enabled = true // figure out where to get this from

		devicecommands = append(devicecommands, dc)

		// add the right things back into d
		d.Commands[index].Endpoint.Name = ep.Name
		d.Commands[index].Endpoint.Path = ep.Path
	}

	// insert everything else
	for _, dr := range deviceroles {
		_, err = accessorGroup.AddDeviceRole(dr)
		if err != nil {
			return Device{}, err
		}
	}

	for _, ps := range devicepowerstates {
		_, err = accessorGroup.AddDevicePowerState(ps)
		if err != nil {
			return Device{}, err
		}
	}

	for _, pc := range portconfigurations {
		_, err = accessorGroup.AddPortConfiguration(pc)
		if err != nil {
			return Device{}, err
		}
	}

	for _, dc := range devicecommands {
		_, err = accessorGroup.AddDeviceCommand(dc)
		if err != nil {
			return Device{}, err
		}
	}

	// clean up d
	d.Room.Devices = nil
	d.Room.Configuration.Evaluators = nil

	return d, nil
}
