package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"net/http"
)

// can i notate this as secret?
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
	req.Header.Add("Authorization", *secret) // I need to set this in the github secrets and pass it here somehow
	if err != nil {
		return err
	}
	return nil
}

var l = log.New(os.Stderr, "", 0)

func main() {
	secret := flag.String("secret", "", "The secret to test")
	err := notify_swarmia(secret)

	if err != nil {
		l.Fatalln(err)
	}
}