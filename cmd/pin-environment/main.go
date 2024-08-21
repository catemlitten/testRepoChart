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


type serviceValues struct {
	ReleaseID string `yaml:"release_id"`
	GlobalReleaseID string `yaml:"global.release_id"`
	GlobalBuildId string `yaml:"global.infra_build_num"`
}

// sets the version inside each service
func setBuildID(ctx context.Context, env, service, buildID string, buildNum string) error {
	dirName := path.Join("state", env, "argo", service)
	fileName := "version.yml"

	fullFileName := path.Join(dirName, fileName)

	// we only want the service directories, not any files in the dir
	fileInfo, err := os.Stat(dirName)
	if !fileInfo.IsDir() {
		return nil
	}
	if err != nil {
		return err
	}

	err = os.MkdirAll(dirName, 0755)
	if err != nil {
		return err
	}

	// Ok so the logic for getting the circleci build number is a different file so doing this for now
	values := serviceValues{ReleaseID: buildID, GlobalReleaseID: buildID, GlobalBuildId: buildNum}

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

// checks if an individual services (ex: doodle) is pinned
func removePins(env string, service string) (bool, error) {
	_, err := os.Stat(filepath.Join("state", env, "argo", service, "pinned"))
	if os.IsNotExist(err) {
		return nil // not existing means not pinned
	}
	if err != nil {
		return false, err
	}
	os.Remove(filepath.Join("state", env, "argo", service, "pinned")) // remove pin
	return nil
}

func mainerr() error {
	ctx := context.Background()

	env := flag.String("env", "", "The environment to pin")
	buildId := flag.String("buildId", "", "The build id to pin to")
	buildNum := flag.String("buildNum", "", "The build num to pin to")

	flag.Parse()

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
        err := removePins(*env, service)
		if err != nil {
            return err
        }
		err = setBuildID(ctx, *env, service, *buildId, *buildNum)
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