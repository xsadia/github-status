package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gen2brain/beeep"
)

const endpoint = "https://www.githubstatus.com/api/v2/status.json"

type GithubStatusData struct {
	Page   Page   `json:"page"`
	Status Status `json:"status"`
}

type Page struct {
	Id        string    `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Status struct {
	Indicator   string `json:"indicator"`
	Description string `json:"description"`
}

func main() {
	client := &http.Client{
		Timeout: time.Second * 5,
	}

	ch := make(chan GithubStatusData)
	go checkStatus(client, endpoint, ch)

	for {
		msg := <-ch
		fmt.Println(msg)
		if msg.Status.Indicator != "none" {
			if err := beeep.Alert("Github: "+msg.Status.Indicator, msg.Status.Description, []byte{}); err != nil {
				log.Println("Notification error: ", err.Error())
			}
		}
	}

}

func checkStatus(client *http.Client, url string, ch chan GithubStatusData) {
	retries := 0
	for {
		if retries >= 5 {
			time.Sleep(time.Minute * 5)
			retries = 0
		}
		resp, err := client.Get(url)
		if err != nil {
			log.Println("Failed to get url ", err.Error())
			retries++
			continue
		}

		retries = 0
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			log.Println("Failed to read body ", err.Error())
			continue
		}

		ghStatusData := GithubStatusData{}
		err = json.Unmarshal(body, &ghStatusData)
		if err != nil {
			log.Println("Failed to unmarshal body ", err.Error())
			continue
		}

		ch <- ghStatusData

		resp.Body.Close()
		time.Sleep(time.Minute * 5)
	}
}
