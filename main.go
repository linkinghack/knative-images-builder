package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	//ServingRootDir         = "./serving"
	//EventingRootDir        = "./eventing"
	//EventingKafkaDir       = "./eventing-kafka"
	//EventingKafkaBrokerDir = "./eventing-kafka-broker"
	//NetIstioDir            = "./net-istio"
	ProjectRootDir = "./serving"
	KoDockerRepo   = ""
	Tags           = "v1.2"
	LocalLoad      = false

	PushLocalImages = false
	TargetRepo      = "linkinghack"
	ReplaceSlash    = false
	RepoUserName    = ""
	RepoPassword    = ""
	Email           = ""
	ServerAddress   = ""
)

func init() {
	//flag.StringVar(&ServingRootDir, "serving-dir", "./serving", "specify root dir of knative serving source codes")
	//flag.StringVar(&EventingRootDir, "eventing-dir", "./eventing", "specify root dir of knative eventing source codes")
	//flag.StringVar(&EventingKafkaDir, "eventing-kafka-dir", "./eventing-kafka", "specify root dir of knative eventing-kafka source codes")
	//flag.StringVar(&EventingKafkaBrokerDir, "eventing-kafka-broker-dir", "./eventing-kafka-broker", "specify root dir of knative eventing-kafka-broker source codes")
	//flag.StringVar(&NetIstioDir, "net-istio-dir", "./net-istio", "specify root dir of knative net-isito source codes")
	flag.StringVar(&ProjectRootDir, "project-root-dir", "./serving", "Root dir of the target knative project dir. Like ./serving, ./eventing etc.")
	flag.StringVar(&KoDockerRepo, "ko-docker-repo", "linkinghack", "auto set KO_DOCKER_REPO env")
	flag.StringVar(&Tags, "tags", "v1.2", "set image tags with ko build --tags")
	flag.BoolVar(&LocalLoad, "local", false, "Load images into local docker daemon")

	flag.BoolVar(&PushLocalImages, "push-local-images", false, "Whether tag and push local ko.local image to specified repo. If specified, build process will be skipped.")
	flag.BoolVar(&ReplaceSlash, "replace-dash", false, "Whether replace slashes (/) with dash (-) in the image names")
	flag.StringVar(&TargetRepo, "target-repo", "linkinghack", "To which repo push the local ko.local images")
	flag.StringVar(&RepoUserName, "repo-user-name", "", "Username of the target repo")
	flag.StringVar(&RepoPassword, "repo-password", "", "Password of the target repo and username")
	flag.StringVar(&Email, "email", "", "Email of the repo user")
	flag.StringVar(&ServerAddress, "server-address", "docker.io", "Server address of the repo without protocol")
}

// ko build ./cmd/controller --platform all --tag-only -B -P --tags 1.2 -L

func main() {
	flag.Parse()
	if PushLocalImages {
		TagAndPushLocalImages()
		return
	}

	fmt.Println(ProjectRootDir)
	KoBuildImages(filepath.Join(ProjectRootDir, "cmd"))
}

var successCmds = []string{}
var failedCmds = []string{}
var succeedImages = []string{}

func KoBuildImages(cmdPath string) {
	if len(cmdPath) < 1 {
		return
	}

	if absPath, err := filepath.Abs(cmdPath); err != nil {
		logrus.Errorf("Cannot recognize file path: %s", cmdPath)
		return
	} else {
		logrus.Infof("Start to traverse path: %s", absPath)
	}

	if err := os.Chdir(cmdPath); err != nil {
		logrus.Errorf("Change dir to %s failed", cmdPath)
	}

	if err := os.Chdir("../"); err != nil {
		logrus.Errorf("Change dir to %s  ../ failed", cmdPath)
	}

	pwd, _ := filepath.Abs(".")
	logrus.Infof("KoBuildImages: current working dir: %s", pwd)

	TraverseDirToBuildMain("./cmd")

	fmt.Println("Succeed commands: ")
	for _, cmd := range successCmds {
		fmt.Println(cmd)
	}

	fmt.Println("Failed commands: ")
	for _, cmd := range failedCmds {
		fmt.Println(cmd)
	}

	fmt.Println("Pushed images: ")
	for _, img := range succeedImages {
		fmt.Println(img)
	}
}

// Traverse and find main.go to execute ko build
//  The file 'path' pointed at should be a dir
func TraverseDirToBuildMain(path string) {
	// Test if it contains main.go.
	//  If there exists main.go file, execute ko build and stop traverse.
	if _, err := os.Open(filepath.Join(path, "main.go")); err != nil {
		logrus.Infof("main.go not found at: %s", path)
	} else {
		logrus.Infof("main.go exists at %s, start Building", path)
		ExecuteKoBuild(path)
		return
	}

	pwd, _ := filepath.Abs(".")
	logrus.Infof("TraverseDirToBuildMain: current working dir: %s", pwd)

	fileEntries, err := os.ReadDir(path)
	if err != nil {
		logrus.Warnf("Cannot open dir: %s, err=%+v,  Skip traverse.", path, err)
		return
	}

	for _, subDir := range fileEntries {
		if subDir.IsDir() {
			logrus.Infof("Traverse path: %s.", subDir.Name())
			TraverseDirToBuildMain("./" + filepath.Join(path, subDir.Name()))
		}
	}
}

func ExecuteKoBuild(path string) {
	currentPath, _ := filepath.Abs(".")
	logrus.Infof("CurrentPath: %s", currentPath)

	// main.go found
	args := []string{
		"build",
		path,
		"--platform",
		"all",
		"--tag-only",
		"-B",
		"-P",
	}
	args = append(args, "--tags", Tags)
	if LocalLoad {
		args = append(args, "-L")
	}

	cmd := exec.Command("ko", args...)
	cmd.Env = append(cmd.Env, os.Environ()...)
	if len(KoDockerRepo) > 0 {
		cmd.Env = append(cmd.Env, "KO_DOCKER_REPO="+KoDockerRepo)
	}

	//cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	output := bytes.Buffer{}
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		logrus.Errorf("Exec %s failed. err=%+v", cmd.String(), err)
		return
	}

	// record executed cmds
	if cmd.ProcessState.Success() {
		successCmds = append(successCmds, cmd.String())

		if imageTag, err := ioutil.ReadAll(&output); err == nil {
			succeedImages = append(succeedImages, string(imageTag))
			logrus.Infof("Image pushed: %s", string(imageTag))
		} else {
			logrus.Warnf("Get ko build STDOUT failed. err=%+v", err)
		}
	} else {
		failedCmds = append(failedCmds, cmd.String())
	}

}
