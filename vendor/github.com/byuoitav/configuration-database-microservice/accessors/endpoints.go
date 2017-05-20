package accessors

import (
	"database/sql"
	"log"
)

func (accessorGroup *AccessorGroup) GetAllEndpoints() ([]Endpoint, error) {

	rows, err := accessorGroup.Database.Query("SELECT * FROM Endpoints")
	if err != nil {
		return []Endpoint{}, err
	}

	endpoints, err := exctractEndpointData(rows)
	if err != nil {
		return []Endpoint{}, err
	}
	defer rows.Close()

	return endpoints, nil
}

func (accessorGroup *AccessorGroup) AddEndpoint(toAdd Endpoint) (Endpoint, error) {

	response, err := accessorGroup.Database.Exec("INSERT INTO Endpoints (name, path, description) VALUES(?,?,?)", toAdd.Name, toAdd.Path, toAdd.Description)
	if err != nil {
		return Endpoint{}, err
	}

	id, err := response.LastInsertId()
	toAdd.ID = int(id)

	return toAdd, nil
}

func (accessorGroup *AccessorGroup) RemoveEndpointByName(name string) error {

	_, err := accessorGroup.Database.Exec("DELETE FROM Endpoints WHERE name=?", name)
	if err != nil {
		return err
	}

	return nil
}

func (accessorGroup *AccessorGroup) GetEndpointByName(name string) (Endpoint, error) {
	row := accessorGroup.Database.QueryRow("SELECT * FROM Endpoints WHERE name = ? ", name)

	e, err := extractEndpoint(row)
	if err != nil {
		return Endpoint{}, err
	}

	return e, nil
}

func exctractEndpointData(rows *sql.Rows) ([]Endpoint, error) {

	var endpoints []Endpoint
	var endpoint Endpoint
	var id *int
	var name *string
	var path *string
	var description *string

	for rows.Next() {
		err := rows.Scan(&id, &name, &path, &description)
		if err != nil {
			return []Endpoint{}, err
		}

		if id != nil {
			endpoint.ID = *id
		}
		if name != nil {
			endpoint.Name = *name
		}
		if path != nil {
			endpoint.Path = *path
		}
		if description != nil {
			endpoint.Description = *description
		}

		endpoints = append(endpoints, endpoint)

	}

	err := rows.Err()
	if err != nil {
		return []Endpoint{}, err
	}

	return endpoints, nil
}

func extractEndpoint(row *sql.Row) (Endpoint, error) {
	var e Endpoint
	var id *int
	var name *string
	var path *string
	var description *string

	err := row.Scan(&id, &name, &path, &description)
	if err != nil {
		log.Printf("error: %s", err.Error())
		return Endpoint{}, err
	}
	if id != nil {
		e.ID = *id
	}
	if name != nil {
		e.Name = *name
	}
	if path != nil {
		e.Path = *path
	}
	if description != nil {
		e.Description = *description
	}

	return e, nil
}
