package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"http"

	"gopkg.in/yaml.v3"
)


func notify_swarmia(secret) nil {
	payload_prep := fmt.Sprintf(`{
        "version": %s, 
        "appName": "hound-fake",
        "environment": %s,
        "repositoryFullName": "honeycombio/fakerepo",
        "commitSha": %s}`, secret, secret, sceret)
	payload := []byte(payload_prep)
	req, err := http.NewRequest("GET", "https://hook.swarmia.com/deployments", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", secret) // I need to set this in the github secrets and pass it here somehow

func main() {
	err := notify_swarmia()

	if err != nil {
		l.Fatalln(err)
	}
}