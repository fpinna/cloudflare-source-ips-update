package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

type cloudflareIps struct {
	Result struct {
		Ipv4Cidrs []string `json:"ipv4_cidrs"`
		Ipv6Cidrs []string `json:"ipv6_cidrs"`
		Etag      string   `json:"etag"`
	} `json:"result"`
	Success  bool  `json:"success"`
	Errors   []any `json:"errors"`
	Messages []any `json:"messages"`
}

func getJsonIps() (cloudflareIps, error) {

	url := "https://api.cloudflare.com/client/v4/ips"

	spaceClient := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := io.ReadAll(res.Body) //FIX-
	if readErr != nil {
		log.Fatal(readErr)
	}

	ips := cloudflareIps{}
	jsonErr := json.Unmarshal(body, &ips)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return ips, err
}
