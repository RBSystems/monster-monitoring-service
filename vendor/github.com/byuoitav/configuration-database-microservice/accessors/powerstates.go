package accessors

import (
	"database/sql"
	"log"
)

type PowerState struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (accessorGroup *AccessorGroup) GetPowerStates() ([]PowerState, error) {
	rows, err := accessorGroup.Database.Query("SELECT * FROM PowerStates")
	if err != nil {
		return []PowerState{}, err
	}

	powerstates, err := extractPowerStates(rows)
	if err != nil {
		return []PowerState{}, err
	}
	defer rows.Close()

	return powerstates, nil
}

func (accessorGroup *AccessorGroup) AddPowerState(powerstate PowerState) (PowerState, error) {
	result, err := accessorGroup.Database.Exec("Insert into PowerStates (powerStateID, name, description) VALUES(?,?,?)", powerstate.ID, powerstate.Name, powerstate.Description)
	if err != nil {
		return PowerState{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return PowerState{}, err
	}

	powerstate.ID = int(id)
	return powerstate, nil
}

func (accessorGroup *AccessorGroup) GetPowerStateByID(id int) (PowerState, error) {
	row := accessorGroup.Database.QueryRow("SELECT * FROM PowerStates WHERE powerStateID = ?", id)

	ps, err := extractPowerState(row)
	if err != nil {
		return PowerState{}, err
	}

	return ps, nil
}

func (accessorGroup *AccessorGroup) GetPowerStateByName(name string) (PowerState, error) {
	row := accessorGroup.Database.QueryRow("SELECT * FROM PowerStates WHERE name = ?", name)

	ps, err := extractPowerState(row)
	if err != nil {
		return PowerState{}, err
	}

	return ps, nil
}

func extractPowerStates(rows *sql.Rows) ([]PowerState, error) {
	var powerstates []PowerState
	var ps PowerState
	var id *int
	var name *string
	var description *string

	for rows.Next() {
		err := rows.Scan(&id, &name, &description)
		if err != nil {
			log.Printf("error: %s", err.Error())
			return []PowerState{}, err
		}
		if id != nil {
			ps.ID = *id
		}
		if name != nil {
			ps.Name = *name
		}
		if description != nil {
			ps.Description = *description
		}

		powerstates = append(powerstates, ps)
	}
	return powerstates, nil
}

func extractPowerState(row *sql.Row) (PowerState, error) {
	var ps PowerState
	var id *int
	var name *string
	var description *string

	err := row.Scan(&id, &name, &description)
	if err != nil {
		log.Printf("error: %s", err.Error())
		return PowerState{}, err
	}
	if id != nil {
		ps.ID = *id
	}
	if name != nil {
		ps.Name = *name
	}
	if description != nil {
		ps.Description = *description
	}

	return ps, nil
}
