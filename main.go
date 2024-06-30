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
	fmt.Println("Automazione della configurazione del server con Go!")

	execDockerCompose()
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

func dockerPs() {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	
	if err != nil {
		panic(err)
	}

	defer apiClient.Close()

	containers, err := apiClient.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		panic(err)
	}

	for _, ctr := range containers {
		fmt.Printf("%s %s (status: %s)\n", ctr.ID, ctr.Image, ctr.Status)
	}
}

func stopAllContainers() {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	defer apiClient.Close()

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


