package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"strings"
)

func TagAndPushLocalImages() {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	imagesPushed := []string{}
	imagesFailed := []string{}

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	for _, image := range images {
		logrus.Infof("Image: Id=%s \n", image.ID)
		for _, imageAndTag := range image.RepoTags {
			nameTag := strings.Split(imageAndTag, ":")
			if len(nameTag) != 2 {
				logrus.Errorf("Cannot recognize RepoTag: %s", imageAndTag)
				continue
			}

			imageName := nameTag[0]
			imageTag := nameTag[1]

			if strings.Contains(imageName, "ko.local/knative.dev") {

				nameParts := strings.Split(imageName, "/")
				if len(nameParts) < 3 {
					logrus.Warnf("Image not recognized: %s. Skip", image.ID)
					continue
				}
				nameParts = nameParts[2:]

				sep := "/"
				if ReplaceSlash {
					sep = "-"
				}

				newImageName := fmt.Sprintf("%s/%s:%s", TargetRepo, strings.Join(nameParts, sep), imageTag)
				logrus.Infof("ko.local image found. Tag and push it: [%s] AS [%s]", image.ID, newImageName)

				cli.ImageTag(context.Background(), image.ID, newImageName)
				if resp, err := cli.ImagePush(context.Background(), newImageName, types.ImagePushOptions{}); err != nil {
					logrus.Warnf("Error push image: %s", newImageName)
					imagesFailed = append(imagesFailed, newImageName)
				} else {
					imagesPushed = append(imagesPushed, newImageName)
					resp.Close()
				}

			}

		}

		fmt.Println("------")
	}

}
