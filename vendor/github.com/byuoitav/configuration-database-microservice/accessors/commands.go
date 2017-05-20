package accessors

import (
	"database/sql"
	"log"
)

/*Command represents all the information needed to issue a particular command to a device.
Name: Command name
Endpoint: the endpoint within the microservice
Microservice: the location of the microservice to call to communicate with the device.
Priority: The relative priority of the command relative to other commands. Commands
					with a higher (closer to 1) priority will be issued to the devices first.
*/
type Command struct {
	Name         string   `json:"name"`
	Endpoint     Endpoint `json:"endpoint"`
	Microservice string   `json:"microservice"`
	Priority     int      `json:"priority"`
}

/*RawCommand represents all the information needed to issue a particular command to a device.
Name: Command name
Description: command description
Priority: The relative priority of the command relative to other commands. Commands
					with a higher (closer to 1) priority will be issued to the devices first.
*/
type RawCommand struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
}

//CommandSorterByPriority sorts commands by priority and implements sort.Interface
type CommandSorterByPriority struct {
	Commands []RawCommand
}

//Len is part of sort.Interface
func (c *CommandSorterByPriority) Len() int {
	return len(c.Commands)
}

//Swap is part of sort.Interface
func (c *CommandSorterByPriority) Swap(i, j int) {
	c.Commands[i], c.Commands[j] = c.Commands[j], c.Commands[i]
}

//Less is part of sort.Interface
func (c *CommandSorterByPriority) Less(i, j int) bool {
	return c.Commands[i].Priority < c.Commands[j].Priority
}

//GetAllCommands simply dumps the commands table
func (accessorGroup *AccessorGroup) GetAllCommands() (commands []RawCommand, err error) {
	log.Printf("Getting all commands...")
	rows, err := accessorGroup.Database.Query("Select * FROM Commands")
	if err != nil {
		log.Printf("Error: %s", err.Error())
		return
	}
	defer rows.Close()

	commands, err = ExtractRawCommands(rows)
	log.Printf("Done.")
	return
}

func (accessorGroup *AccessorGroup) GetRawCommandByName(name string) (RawCommand, error) {
	row := accessorGroup.Database.QueryRow("SELECT * FROM Commands WHERE name = ? ", name)

	rc, err := extractRawCommand(row)
	if err != nil {
		return RawCommand{}, err
	}

	return rc, nil
}

//ExtractCommand pulls a command object from a set of sql.Rows
func ExtractCommand(rows *sql.Rows) (allCommands []Command, err error) {

	for rows.Next() {
		command := Command{}

		err = rows.Scan(&command.Name, &command.Endpoint.Name, &command.Endpoint.Path, &command.Microservice)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			return
		}

		allCommands = append(allCommands, command)
	}

	return
}

//ExtractRawCommands pulls a RawCommand object from a set of sql.Rows
func ExtractRawCommands(rows *sql.Rows) (allCommands []RawCommand, err error) {

	for rows.Next() {
		command := RawCommand{}

		err = rows.Scan(&command.ID, &command.Name, &command.Description, &command.Priority)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			return
		}

		allCommands = append(allCommands, command)
	}

	return
}

func (accessorGroup *AccessorGroup) AddRawCommand(rc RawCommand) (RawCommand, error) {
	result, err := accessorGroup.Database.Exec("Insert into Commands (commandID, name, description, priority) VALUES(?,?,?,?)", rc.ID, rc.Name, rc.Description, rc.Priority)
	if err != nil {
		return RawCommand{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return RawCommand{}, err
	}

	rc.ID = int(id)
	return rc, nil
}

func extractRawCommand(row *sql.Row) (RawCommand, error) {
	var rc RawCommand
	var id *int
	var name *string
	var description *string
	var priority *int

	err := row.Scan(&id, &name, &description, &priority)
	if err != nil {
		log.Printf("error: %s", err.Error())
		return RawCommand{}, err
	}
	if id != nil {
		rc.ID = *id
	}
	if name != nil {
		rc.Name = *name
	}
	if description != nil {
		rc.Description = *description
	}
	if priority != nil {
		rc.Priority = *priority
	}

	return rc, nil
}
