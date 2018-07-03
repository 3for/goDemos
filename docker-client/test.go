package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/fsouza/go-dockerclient"
)

func main() {
	//endpoint := "unix:///var/run/docker.sock"
	endpoint := "192.168.1.248:5555"
	client, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}
	/* imgs, err := client.ListImages(docker.ListImagesOptions{All: false})
	if err != nil {
		panic(err)
	}
	for _, img := range imgs {
		fmt.Println("ID: ", img.ID)
		fmt.Println("RepoTags: ", img.RepoTags)
		fmt.Println("Created: ", img.Created)
		fmt.Println("Size: ", img.Size)
		fmt.Println("VirtualSize: ", img.VirtualSize)
		fmt.Println("ParentId: ", img.ParentID)
	} */

	//containers, err := client.ListContainers(docker.ListContainersOptions{All: false})
	containerName := "/event-node1"
	containers, err := client.ListContainers(docker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			//"ancestor": []string{"hyperledger/fabric-peer"}, //filter containers by specific image.
			"name": []string{containerName},
		},
	})

	if err != nil {
		panic(err)
	}
	for _, ctr := range containers {
		fmt.Println("container ID: ", ctr.ID)

		client.Logs

		client.SkipServerVersionCheck = true
		// Reading logs from container a84849 and sending them to buf.
		var buf bytes.Buffer
		err = client.AttachToContainer(docker.AttachToContainerOptions{
			Container:    ctr.ID,
			OutputStream: &buf,
			Logs:         true,
			Stdout:       true,
			Stderr:       true,
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Println(buf.String())
		buf.Reset()
		err = client.AttachToContainer(docker.AttachToContainerOptions{
			Container:    "a84849",
			OutputStream: &buf,
			Stdout:       true,
			Stream:       true,
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Println(buf.String())

		/* fmt.Println("container Image: ", ctr.Image)
		fmt.Println("container Command: ", ctr.Command)
		fmt.Println("container Created: ", ctr.Created)
		fmt.Println("container State: ", ctr.State)
		fmt.Println("container Status: ", ctr.Status)
		fmt.Println("container Ports: ", ctr.Ports)
		fmt.Println("container SizeRw: ", ctr.SizeRw)
		fmt.Println("container SizeRootFs: ", ctr.SizeRootFs)
		fmt.Println("container Names: ", ctr.Names)
		fmt.Println("container Labels: ", ctr.Labels)

		fmt.Println("container Networks: ", ctr.Networks)
		fmt.Println("container Mounts: ", ctr.Mounts) */
	}
}
