package accessors

import (
	"database/sql"
	"log"
)

type DeviceRoleDef struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (accessorGroup *AccessorGroup) GetDeviceRoleDefs() ([]DeviceRoleDef, error) {
	rows, err := accessorGroup.Database.Query("SELECT * FROM DeviceRoleDefinition")
	if err != nil {
		return []DeviceRoleDef{}, err
	}

	deviceroledefs, err := extractDeviceRoleDefs(rows)
	if err != nil {
		return []DeviceRoleDef{}, err
	}
	defer rows.Close()

	return deviceroledefs, nil
}

func (accessorGroup *AccessorGroup) AddDeviceRoleDef(deviceroledef DeviceRoleDef) (DeviceRoleDef, error) {
	result, err := accessorGroup.Database.Exec("Insert into DeviceRoleDefinition (deviceRoleDefinitionID, name, description) VALUES(?,?,?)", deviceroledef.ID, deviceroledef.Name, deviceroledef.Description)
	if err != nil {
		return DeviceRoleDef{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return DeviceRoleDef{}, err
	}

	deviceroledef.ID = int(id)
	return deviceroledef, nil
}

func (accessorGroup *AccessorGroup) GetDeviceRoleDefByID(id int) (DeviceRoleDef, error) {
	row := accessorGroup.Database.QueryRow("SELECT * FROM DeviceRoleDefinition WHERE deviceRoleDefinitionID = ? ", id)

	drd, err := extractDeviceRoleDef(row)
	if err != nil {
		return DeviceRoleDef{}, err
	}

	return drd, nil
}

func (accessorGroup *AccessorGroup) GetDeviceRoleDefByName(name string) (DeviceRoleDef, error) {
	row := accessorGroup.Database.QueryRow("SELECT * FROM DeviceRoleDefinition WHERE name = ? ", name)

	drd, err := extractDeviceRoleDef(row)
	if err != nil {
		return DeviceRoleDef{}, err
	}

	return drd, nil
}

func extractDeviceRoleDefs(rows *sql.Rows) ([]DeviceRoleDef, error) {
	var deviceroledefs []DeviceRoleDef
	var deviceroledef DeviceRoleDef
	var id *int
	var name *string
	var description *string

	for rows.Next() {
		err := rows.Scan(&id, &name, &description)
		if err != nil {
			log.Printf("error: %s", err.Error())
			return []DeviceRoleDef{}, err
		}
		if id != nil {
			deviceroledef.ID = *id
		}
		if name != nil {
			deviceroledef.Name = *name
		}
		if description != nil {
			deviceroledef.Description = *description
		}

		deviceroledefs = append(deviceroledefs, deviceroledef)
	}
	return deviceroledefs, nil
}

func extractDeviceRoleDef(row *sql.Row) (DeviceRoleDef, error) {
	var drd DeviceRoleDef
	var id *int
	var name *string
	var description *string

	err := row.Scan(&id, &name, &description)
	if err != nil {
		log.Printf("error: %s", err.Error())
		return DeviceRoleDef{}, err
	}
	if id != nil {
		drd.ID = *id
	}
	if name != nil {
		drd.Name = *name
	}
	if description != nil {
		drd.Description = *description
	}

	return drd, nil
}
