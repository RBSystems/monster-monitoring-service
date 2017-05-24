package salt

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

type SaltConnection struct {
	Token    string
	Expires  float64
	Response *http.Response
}

type LoginResponse struct {
	Eauth       string   `json:"eauth,omitempty"`
	Expire      float64  `json:"expire,omitempty"`
	Permissions []string `json:"perms,omitempty"`
	Start       float64  `json:"start,omitempty"`
	Token       string   `json:"token,omitempty"`
	User        string   `json:"user,omitempty"`
}

type SaltEvent struct {
	Tag  string                 `json:"tag"`
	Data map[string]interface{} `json:"data"`
}

//singleton instance of salt connection
var connection *SaltConnection
var once sync.Once

func Connection() *SaltConnection {
	once.Do(func() {
		log.Printf("Logging into salt...")
		connection.Login()
	})
	return connection
}

func (sc *SaltConnection) Login() error {
	log.Printf("Logging into the salt master")

	values := make(map[string]string)
	values["username"] = os.Getenv("SALT_EVENT_USERNAME")
	values["password"] = os.Getenv("SALT_EVENT_PASSWORD")
	values["eauth"] = "pam"

	b, _ := json.Marshal(values)

	req, err := http.NewRequest("POST", os.Getenv("SALT_MASTER_ADDRESS")+"/login", bytes.NewBuffer(b))
	if err != nil {
		log.Printf("Error building the request: %s", err.Error())
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	log.Printf("Headers set, sending request")

	//For now ignore the certificate error, eventually we'll need to get one
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending the login request: %s", err.Error())
		return err
	}

	log.Printf("Request sent")

	respBody := make(map[string][]LoginResponse)

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading the login response: %s", err.Error())
	}

	log.Printf("Body: %s", b)
	err = json.Unmarshal(b, &respBody)
	if err != nil {
		log.Printf("Error unmarshalling login response: %s", err.Error())
	}

	log.Printf("Struct %+v", respBody)
	lr := respBody["return"][0]

	sc.Token = lr.Token
	sc.Expires = lr.Expire
	log.Printf("Done.")
	return nil
}
