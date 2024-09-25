package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"time"
	"net"
	"net/http"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   7 * time.Second,
		ResponseHeaderTimeout: 9 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,
	},
	Timeout: 12 * time.Second,
} //borrowed

func notify_swarmia(secret *string) error {
	payload_prep := fmt.Sprintf(`{
        "version": %s, 
        "appName": "hound-fake",
        "environment": %s,
        "repositoryFullName": "honeycombio/fakerepo",
        "commitSha": %s}`, secret, secret, secret)
	payload := []byte(payload_prep)
	req, err := http.NewRequest("GET", "https://hook.swarmia.com/deployments", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", *secret)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.StatusCode)
		fmt.Println(string(body))
		return nil
	}
	return nil
}

var l = log.New(os.Stderr, "", 0)

func main() {
	// secret := flag.String("secret", "", "The secret to test")
	// err := notify_swarmia(secret)
	for _, env := range os.Environ() {
		println(env)
	}
}