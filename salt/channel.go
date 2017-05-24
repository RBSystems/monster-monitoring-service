package salt

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func Listen() {

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

	reader := bufio.NewReader(Connection().Response.Body)
	for {

		if Connection().Response.Close {
			log.Printf("Detected closed salt connection. Exiting...")
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

					//					err = store.UpdateStoreBySalt(event)
					if err != nil {
						log.Printf("Error writing to badger store: %s", err.Error())
					}

				}
			} else if len(line) < 1 {
				continue
			}
		}
	}
}

func Start(timer chan bool) {

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)

	go Wait(signals, timer)
	<-timer
	log.Printf("SIGTERM interrupt detected. Exiting...")

	//store.Store().Close()
	//(*Connection().Connection).Close()
	Connection().Response.Body.Close()
	os.Exit(0)

}

func Wait(signals chan os.Signal, timer chan bool) {

	<-signals
	timer <- true
}
