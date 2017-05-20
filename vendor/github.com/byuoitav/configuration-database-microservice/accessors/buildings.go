package accessors

type Building struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Shortname   string `json:"shortname,omitempty"`
	Description string `json:"description,omitempty"`
}

// GetAllBuildings returns a list of buildings from the database
func (accessorGroup *AccessorGroup) GetAllBuildings() ([]Building, error) {
	allBuildings := []Building{}

	rows, err := accessorGroup.Database.Query("SELECT * FROM Buildings")
	if err != nil {
		return []Building{}, err
	}

	defer rows.Close()

	for rows.Next() {
		building := Building{}

		err = rows.Scan(&building.ID, &building.Name, &building.Shortname, &building.Description)
		if err != nil {
			return []Building{}, err
		}

		allBuildings = append(allBuildings, building)
	}

	err = rows.Err()
	if err != nil {
		return []Building{}, err
	}

	return allBuildings, nil
}

// GetBuildingByID returns a building from the database by ID
func (accessorGroup *AccessorGroup) GetBuildingByID(id int) (Building, error) {
	building := &Building{}
	err := accessorGroup.Database.QueryRow("SELECT * FROM Buildings WHERE buildingID=?", id).Scan(&building.ID, &building.Name, &building.Shortname, &building.Description)
	if err != nil {
		return Building{}, err
	}

	return *building, nil
}

// GetBuildingByShortname returns a building from the database by shortname
func (accessorGroup *AccessorGroup) GetBuildingByShortname(shortname string) (Building, error) {
	building := &Building{}
	err := accessorGroup.Database.QueryRow("SELECT * FROM Buildings WHERE shortname=?", shortname).Scan(&building.ID, &building.Name, &building.Shortname, &building.Description)
	if err != nil {
		return Building{}, err
	}

	return *building, nil
}

//AddBuilding adds a building

func (accessorGroup *AccessorGroup) AddBuilding(name string, shortname string, description string) (Building, error) {

	result, err := accessorGroup.Database.Exec(`INSERT into Buildings (name, shortname, description) VALUES (?,?,?)`, name, shortname, description)
	if err != nil {
		return Building{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Building{}, err
	}

	building := Building{
		Name:        name,
		Shortname:   shortname,
		Description: description,
	}
	building.ID = int(id) // cast id into an int

	return building, err
}
