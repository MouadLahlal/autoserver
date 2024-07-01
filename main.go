package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"log"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func main() {
	PrepareNpm()
	PrepareJellyfin()
	StartContainers()
}

func PrepareNpm() {
	cmd := exec.Command("docker", "network", "create", "common-npm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("Error in creating docker network : %s", err)
	}

	mkdir("/srv/npm/data")
	mkdir("/srv/npm/letsencrypt")
}

func PrepareJellyfin() {
	mkdir("/srv/jellyfin/config")
	mkdir("/srv/jellyfin/cache")
	mkdir("/media/movies")
	mkdir("/media/series")
	mkdir("/media/music")
}

func StartContainers() {
	dir := "docker-compose"
	
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("Directory", dir, "does not exist")
		return
	}

	files, err := os.ReadDir("docker-compose")
	
	if err != nil {
		log.Fatal(err)
	}
	
	for _, file := range files {
		if !(file.IsDir()) {
			fmt.Println("Found service :", file.Name())
			cmd := exec.Command("docker", "compose", "-f", file.Name(), "up", "-d")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Dir = "docker-compose"

			if err := cmd.Run(); err != nil {
				log.Fatalf("Error in starting service : %s", err)
			}
		}
	}
}

func execDockerCompose() {
	if _, err := os.Stat("docker-compose.yml"); os.IsNotExist(err) {
		log.Fatalf("docker-compose.yml file not found")
	}

	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Running docker compose")

	if err := cmd.Run(); err != nil {
		log.Fatalf("Error in executing docker compose: %s", err)
	}

	fmt.Println("Container/s successfully created")
}

func stopAllContainers() {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	defer apiClient.Close()

	os.Setenv("DOCKER_API_VERSION", "1.45")
	fmt.Println(os.Getenv("DOCKER_API_VERSION"))
	fmt.Println(os.Getenv("HOME"))

	containers, err := apiClient.ContainerList(context.Background(), container.ListOptions{All: false})
	if err != nil {
		panic(err)
	}

	for _, ctr := range containers {
		fmt.Printf("ID: %s \nName: %s \n\n", ctr.ID, ctr.Image)
		err := apiClient.ContainerStop(context.Background(), ctr.ID, container.StopOptions{Timeout: nil, Signal: ""})
		if err != nil {
			panic(err)
		}
	}
}

func mkdir(dirPath string) {
	err := os.MkdirAll(dirPath, 0755)

	if err != nil {
		fmt.Printf("Errore nella creazone della directory: %v\n", err)
	} else {
		fmt.Printf("Directory %s creata con successo.\n", dirPath)
	}
}


