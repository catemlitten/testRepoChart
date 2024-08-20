package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var l = log.New(os.Stderr, "", 0)
var services []string

// latest-build-id lives at the root of the kubernetes dir, outside of all the envs and their services
func getLatestBuildID(ctx context.Context) (string, error) {
	read, err := os.ReadFile("state/latest-build-id")

	if err != nil {
		return "", err
	}

	stringified := string(read)

	stripped := strings.Trim(stringified, "\n")

	return stripped, nil
}

type serviceValues struct {
	ReleaseID string `yaml:"release_id"`
}

// sets the version inside each service
func setBuildID(ctx context.Context, env, service, buildID string) error {
	dirName := path.Join("state", env, "argo", service)
	fileName := "version.yml"

	fullFileName := path.Join(dirName, fileName)

	err := os.MkdirAll(dirName, 0755)
	if err != nil {
		return err
	}

	values := serviceValues{ReleaseID: buildID}

	serialized, err := yaml.Marshal(&values)
	if err != nil {
		return err
	}

	err = os.WriteFile(fullFileName, serialized, 0755)
	if err != nil {
		return err
	}

	return nil

}

// checks if the overall env (ex: staging) is pinned
func isEnvPinned(env string) (bool, error) {
	_, err := os.Stat(filepath.Join("state", env, "pinned"))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// checks if an individual services (ex: email) is pinned
func isServicePinned(env string, service string) (bool, error) {
	_, err := os.Stat(filepath.Join("state", env, "argo", service, "pinned"))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func mainerr() error {
	ctx := context.Background()

	env := flag.String("env", "", "The environment to promote")

	flag.Parse()

	latestBuildID, err := getLatestBuildID(ctx)
	if err != nil {
		return err
	}

	l.Printf("Got %s for latest build id\n", latestBuildID)

    // Check if whole environment is pinned right now
    pinned, err := isEnvPinned(*env)

	if err != nil {
		return err
	}

	if pinned {
        fmt.Printf("%v is pinned", *env) 
		return nil
	}

	err = setBuildID(ctx, *env, "", latestBuildID)
	if err != nil {
		return err
	}

    // Get a list of services from to update by listing the dirs under the regional env
	entries, err := os.ReadDir("state/staging/argo")
    if err != nil {
        log.Fatal(err)
    }
 
    for _, e := range entries {
            fmt.Println(e.Name()) // debug, can be removed?
			services = append(services, e.Name())
    }

    for _, service := range services {
        pinned, err := isServicePinned(*env, service)
        if pinned {
            fmt.Printf("%v is pinned", service) 
            return nil
        }
        err = setBuildID(ctx, *env, service, latestBuildID)
        if err != nil {
            return err
        }
    }

	return nil
}

func main() {
	err := mainerr()

	if err != nil {
		l.Fatalln(err)
	}
}