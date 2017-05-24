package salt

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

func Listen(events chan SaltEvent, done chan bool, signal sync.WaitGroup) {

	var once *sync.Once

	for {
		select {
		case <-done:
			log.Printf("SIGTERM interrupt detected. Closing connection to salt...")
			Connection().Response.Body.Close()
			break
		default:
			once.Do(connect)
			listenSalt(events)
		}
	}
	signal.Done()
}

var reader *bufio.Reader

func connect() {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}

	req, err := http.NewRequest("GET", os.Getenv("SALT_MASTER_ADDRESS")+"/events", nil)
	if err != nil {
		log.Printf("Cannot open request %s", err.Error())
		return
	}

	req.Header.Add("X-Auth-Token", Connection().Token)

	Connection().Response, err = client.Do(req)
	if err != nil {
		log.Printf("Error sending Request %s", err.Error())
		return
	}

	reader = bufio.NewReader(Connection().Response.Body)
}

func listenSalt(events chan SaltEvent) {

	if Connection().Response.Close {
		log.Printf("Detected closed salt connection. Reconnecting...")
		connect()
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Error reading event" + err.Error())
	} else {
		if strings.Contains(line, "retry") {
			return
		} else if strings.Contains(line, "tag") {

			line2, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}

			if strings.Contains(line2, "data") {

				jsonString := line2[5:]
				var event SaltEvent

				err := json.Unmarshal([]byte(jsonString), &event)
				if err != nil {
					log.Fatal("Error unmarshalling event" + err.Error())
				}
				events <- event
			}
		} else if len(line) < 1 {
			return
		}
	}
}
