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

	log.Printf("Starting salt routine...")

	//	var read, listener *sync.Once
	var read, listen, close sync.Once

	for {
		select {
		case <-done:
			close.Do(func() {
				log.Printf("SIGTERM signal detected. Closing connection to salt...")
				Connection().Response.Body.Close()
			})
			break
		default:
			read.Do(func() {
				log.Printf("Please don't panic")
				connect()
			})
			listen.Do(func() {
				go listenSalt(events)
			})
		}
	}
	signal.Done()
}

var reader *bufio.Reader

func connect() {

	log.Printf("Subscribing to salt...")

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

	log.Printf("Reading salt events...")

	for {
		if reader == nil {
			log.Printf("Reader not initialized. Waiting for connection...")
			continue
		}

		if Connection().Response.Close {
			log.Printf("Detected closed salt connection. Terminating process...")
			break
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Error reading event" + err.Error())
		} else {
			if strings.Contains(line, "retry") {
				continue
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
				continue
			}
		}
	}
	return
}
