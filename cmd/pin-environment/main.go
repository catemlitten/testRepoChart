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
var env_type := ""

type serviceValues struct {
	ReleaseID string `yaml:"release_id"`
	GlobalReleaseID string `yaml:"global.release_id"`
	GlobalBuildId string `yaml:"global.infra_build_num"`
}

func getEnvType(env) string {
	if strings.HasPrefix(env, "kibble") {
		env_type = "kibbles"
	} else if strings.HasPrefix(env, "dogfood") {
		env_type = "dogfoods"
	} else if strings.HasPrefix(env, "production") {
		env_type = "productions"
	} else {
		return ""
	}
	return env_type
}

// sets the version inside each service
func setBuildID(ctx context.Context, env, env_type, service, buildID string, buildNum string) error {
	dirName := path.Join("argo-kubernetes-charts", service, "environment_values", env_type)
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

// checks if an individual services (ex: doodle) is pinned and removes it
func removePins(env string, service string) error {
	path, err := os.Stat(filepath.Join("state", env, "argo", service))
	if !path.IsDir() {
		return nil // not a service directory
	}
	if err != nil {
		return err
	}
	_, err = os.Stat(filepath.Join("state", env, "argo", service, "pinned"))
	if os.IsNotExist(err) {
		return nil // not existing means not pinned
	}
	if err != nil {
		return err
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
	// Ok so I have envs per service so I need to iterate through each service and into it's correct folder then drop a pin in that
	entries, err := os.ReadDir("argo-kubernetes-charts")
    if err != nil {
        log.Fatal(err)
    }
 
    for _, e := range entries {
            fmt.Println(e.Name()) // debug, can be removed?
			services = append(services, e.Name())
    }

	// set which env dir (ex: dogfoods) it will go into
	env_type = getEnvType(env)

	// set build id on the services within the env
	for _, service := range services {
        err := removePins(*env, service)
		if err != nil {
            return err
        }
		err = setBuildID(ctx, *env, env_type, service, *buildId, *buildNum)
        if err != nil {
            return err
        }
    }
	// set build id on the env itself
	err = setBuildID(ctx, *env, "", "", *buildId, *buildNum)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	err := mainerr()

	if err != nil {
		l.Fatalln(err)
	}
}