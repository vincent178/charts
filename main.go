package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"

	"gopkg.in/yaml.v3"
)

var queue = map[string]string{}
var dockerHubAccount = "vincent178"
var imageWithSha = regexp.MustCompile(`^gcr\.io.*\/(.*)@sha256:\w{64}$`)
var imageWithTag = regexp.MustCompile(`^gcr\.io.*\/(.*):(\w*)$`)

type chart struct {
	Name       string `yaml:"name"`
	AppVersion string `yaml:"appVersion"`
}

func main() {

	valuesdata, err := ioutil.ReadFile("./tekton/values.yaml")

	if err != nil {
		log.Fatal(fmt.Sprintf("can't read file %s", "./tekton/values.yaml"))
	}

	chartdata, err := ioutil.ReadFile("./tekton/Chart.yaml")

	if err != nil {
		log.Fatal(fmt.Sprintf("can't read file %s", "./tekton/Chart.yaml"))
	}

	c := &chart{}

	node := &yaml.Node{}

	yaml.Unmarshal(valuesdata, node)
	yaml.Unmarshal(chartdata, c)

	walkYaml(node, c)

	valuesdata, err = yaml.Marshal(node)

	if err != nil {
		log.Fatal(err)
	}

	ioutil.WriteFile("./tekton/values.yaml", valuesdata, 0644)

	run("git", "add", ".")
	run("git", "commit", "-m", "update gcr images to docker hub")
	run("git", "push", "origin", "master")
}

// walkYaml will iterate yaml file and do followings
// 1. find gcr.io image
// 2. pull down
// 3. tag it
// 4. push to docker hub
// 5. update yaml
func walkYaml(node *yaml.Node, c *chart) {
	for _, n := range node.Content {
		imagePath := n.Value

		segs, _ := parseImagePath(imagePath)

		if len(segs) == 0 {
			walkYaml(n, c)
			continue
		}

		// this should not happend
		if len(segs) == 1 {
			log.Fatal(errors.New("invalid state"))
		}

		// image is sha256 or latest tag
		// use helm chart appVersion as tag name
		if len(segs) == 2 {
			segs = append(segs, c.AppVersion)
		}

		pushImagePath := fmt.Sprintf("%s/%s-%s:%s", dockerHubAccount, c.Name, segs[1], segs[2])

		pullAndPush(imagePath, pushImagePath)

		n.Value = pushImagePath
	}
}

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
	run("docker", "pull", pullImagePath)
	run("docker", "tag", pullImagePath, pushImagePath)
	run("docker", "push", pushImagePath)
}

func run(name string, arg ...string) {
	cmd := exec.Command(name, arg...)

	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
