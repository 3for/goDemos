package main

import (
	"fmt"
	"os"

	"github.com/fsouza/go-dockerclient"
	"github.com/hpcloud/tail"
)

func main() {
	endpoint := "unix:///var/run/docker.sock"
	//endpoint := "192.168.1.248:5555"
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
	//containerName := "/peer0.org1.example.com"
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

	config := tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}, //tail from the last Nth location
		MustExist: false,
		Poll:      true,
	}
	dockerDirectory := "/var/lib/docker/containers/"

	done := make(chan bool)

	for _, ctr := range containers {
		fmt.Println("container ID: ", ctr.ID)
		filename := dockerDirectory + ctr.ID + "/" + ctr.ID + "-json.log"
		//filename := "/tmp/3709f32a812d5471ac1f54d5cee9e8df7f55102ce327327ba7d810ec88122bbb-json.log"
		go tailFile(filename, config, done)
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

	for _ = range containers {
		<-done
	}
}
func tailFile(filename string, config tail.Config, done chan bool) {
	defer func() { done <- true }()
	t, err := tail.TailFile(filename, config)
	if err != nil {
		fmt.Println(err)
		return
	}
	for line := range t.Lines {
		fmt.Println(line.Text)
	}
	err = t.Wait()
	if err != nil {
		fmt.Println(err)
	}
	/* var msg *tail.Line
	var ok bool
	for true {
		msg, ok = <-t.Lines
		if !ok {
			fmt.Printf("tail file close reopen, filename:%s\n", t.Filename)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		fmt.Println("msg:", msg)
		fmt.Println("msg text:", msg.Text)
	} */
}
