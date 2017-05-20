package accessors

import "database/sql"

type DeviceRole struct {
	ID                     int `json:"id,omitempty"`
	DeviceID               int `json:"device"`
	DeviceRoleDefinitionID int `json:"role"`
}

func (accessorGroup *AccessorGroup) GetDeviceRoles() ([]DeviceRole, error) {
	rows, err := accessorGroup.Database.Query("SELECT * FROM DeviceRole")
	if err != nil {
		return []DeviceRole{}, err
	}

	deviceroles, err := exctractDeviceRoleData(rows)
	if err != nil {
		return []DeviceRole{}, err
	}
	defer rows.Close()

	return deviceroles, nil
}

func (accessorGroup *AccessorGroup) AddDeviceRole(dr DeviceRole) (DeviceRole, error) {
	response, err := accessorGroup.Database.Exec("INSERT INTO DeviceRole (deviceRoleID, deviceID, deviceRoleDefinitionID) VALUES(?,?,?)", dr.ID, dr.DeviceID, dr.DeviceRoleDefinitionID)
	if err != nil {
		return DeviceRole{}, err
	}

	id, err := response.LastInsertId()
	dr.ID = int(id)

	return dr, nil
}

func exctractDeviceRoleData(rows *sql.Rows) ([]DeviceRole, error) {
	var deviceroles []DeviceRole
	var devicerole DeviceRole
	var id *int
	var dID *int
	var rID *int

	for rows.Next() {
		err := rows.Scan(&id, &dID, &rID)
		if err != nil {
			return []DeviceRole{}, err
		}

		if id != nil {
			devicerole.ID = *id
		}
		if dID != nil {
			devicerole.DeviceID = *dID
		}
		if rID != nil {
			devicerole.DeviceRoleDefinitionID = *rID
		}
		deviceroles = append(deviceroles, devicerole)
	}

	err := rows.Err()
	if err != nil {
		return []DeviceRole{}, err
	}

	return deviceroles, nil
}
