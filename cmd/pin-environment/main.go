package main

import (
	"context"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path"
	"strings"
)

var l = log.New(os.Stderr, "", 0)
var services []string
var envType = ""

type serviceValues struct {
	ReleaseID       string `yaml:"release_id"`
	GlobalReleaseID string `yaml:"global.release_id"`
	GlobalBuildId   string `yaml:"global.infra_build_num"`
}

func getEnvType(ctx context.Context, env string) string {
	if strings.HasPrefix(env, "pres") {
		envType = "prestaging"
	} else if strings.HasPrefix(env, "stag") {
		envType = "staging"
	} else if strings.HasPrefix(env, "production") {
		envType = "productions"
	} else {
		return ""
	}
	return envType
}

// sets the version inside each service
func setBuildID(ctx context.Context, env string, envType, service, buildID string, buildNum string) error {
	dirName := path.Join("argo-kubernetes-charts", service, "environment_values", envType)
	fileName := env + "_" + service + "_values.yaml"

	fullFileName := path.Join(dirName, fileName)
	fmt.Println(fullFileName)

	// we only want the service directories, not any files in the dir
	fileInfo, err := os.Stat(dirName)
	if !fileInfo.IsDir() {
		return nil
	}
	if err != nil {
		return err
	}

	// Ok so the logic for getting the circleci build number is a different file so doing this for now
	//values := serviceValues{ReleaseID: buildID, GlobalReleaseID: buildID, GlobalBuildId: buildNum}

	/***
	release_id: "branch-03b75f0ebe66fe38a3f50634e129ff7945848d62"
	global:
	release_id: "branch-03b75f0ebe66fe38a3f50634e129ff7945848d62"
	infra_build_num: "1489606"
	***/

	type Global struct {
		ReleaseId     string `yaml:"release_id"`
		InfraBuildNum string `yaml:"infra_build_num"`
	}

	type Release struct {
		ReleaseId string `yaml:"release_id"`
		Globals   Global `yaml:"global"`
	}

	yamlFile := Release{
		ReleaseId: buildID,
		Globals: Global{
			ReleaseId:     buildID,
			InfraBuildNum: buildNum,
		},
	}

	serialized, err := yaml.Marshal(&yamlFile)
	if err != nil {
		return err
	}

	// can I read the value of the string to replace using yaml unmarshal
	// then use string replace to update it?
	// fuck it, use bash after the fact to add a comment at the top.

	fmt.Printf("writing inside of setbuildid %s", fullFileName)
	err = os.WriteFile(fullFileName, serialized, 0755)
	if err != nil {
		return err
	}

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
	dir := "argo-kubernetes-charts"

	// read the files inside of argo-kubernetes-charts
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf(err.Error())
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue // skip anything not a dir
		}
		if e.Name() == dir {
			continue // skip writing a values file to the root here
		}
		fmt.Printf("appending %s to services \n", e.Name())
		services = append(services, e.Name())
	}

	// set which env dir (ex: dogfoods) it will go into
	envType = getEnvType(ctx, *env)

	// set build id on the services within the env
	for _, service := range services {
		err = setBuildID(ctx, *env, envType, service, *buildId, *buildNum)
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
