package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"

	"gopkg.in/yaml.v2"
)

var dockerHubAccount = "vincent178"
var imageWithSha = regexp.MustCompile(`^gcr\.io.*\/(.*)@sha256:\w{64}$`)
var imageWithTag = regexp.MustCompile(`^gcr\.io.*\/(.*):(\w*)$`)

type values struct {
	Images map[string]string `yaml:"images"`
}

type chart struct {
	Name       string `yaml:"name"`
	AppVersion string `yaml:"appVersion"`
}

func main() {
	valuesdata, err := ioutil.ReadFile("./tekton/values.yaml")

	if err != nil {
		log.Fatal(err)
	}

	v := values{}

	err = yaml.Unmarshal(valuesdata, &v)

	if err != nil {
		log.Fatal(err)
	}

	chartvalue, err := ioutil.ReadFile("./tekton/Chart.yaml")

	c := chart{}

	err = yaml.Unmarshal(chartvalue, &c)

	if err != nil {
		log.Fatal(err)
	}

	for _, imagePath := range v.Images {
		segs, _ := parseImagePath(imagePath)

		if len(segs) == 2 {
			segs = append(segs, c.AppVersion)
		}

		pushImagePath := fmt.Sprintf("%s/%s-%s:%s", dockerHubAccount, c.Name, segs[1], segs[2])
		fmt.Println(pushImagePath)

		// cmd := exec.Command("docker", "push", imagePath)
		// cmd.Stdout = os.Stdout
		// cmd.Run()
	}
}

func isValidImagePath(imagePath string) bool {
	return true
}

// return last segment of path and tag
func parseImagePath(imagePath string) ([]string, error) {
	if imageWithSha.MatchString(imagePath) {
		return imageWithSha.FindStringSubmatch(imagePath), nil
	}

	if imageWithTag.MatchString(imagePath) {
		return imageWithTag.FindStringSubmatch(imagePath), nil
	}

	return nil, errors.New("invalid image path, " + imagePath)
}

func pullAndPush(pullImagePath, pushImagePath string) {
}
