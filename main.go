package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"log"
	"bytes"
	"strings"
	"encoding/json"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"
)

func main() {
	//PrepareNpm()
	//PrepareJellyfin()
	//StartContainers()

	app := tview.NewApplication()
	layout := tview.NewFlex()
	main := tview.NewFlex()

	currentContainer := ""

	containerView := tview.NewList().ShowSecondaryText(false)
	infoView := tview.NewForm()
	shellView := tview.NewFlex()
	statsView := tview.NewForm()

	app.EnableMouse(true)

	shellOutputView := tview.NewTextView()
	shellCommandInput := tview.NewInputField()

	allContainers := GetAllContainers()
	for _, ctr := range allContainers {
		txt := ""
		if ctr.State == "running" {
			txt = fmt.Sprintf("ON - %s", ctr.Image)
		} else {
			txt = fmt.Sprintf("OFF - %s", ctr.Image)
		}
		containerView.AddItem(txt, "", 0, func() {
			currentContainer = ctr.Names[0]
			
			infoView.Clear(true)
			infoView.AddTextView("Image", ctr.Image, 40, 2, true, false)
			infoView.AddTextView("Name", ctr.Names[0], 40, 2, true, false)
			infoView.AddTextView("State", ctr.State, 40, 2, true, false)
			infoView.AddTextView("Status", ctr.Status, 40, 2, true, false)

			statsView.Clear(true)
			statsView.AddTextView("CPU Usage", fmt.Sprintf("%.2f%%", GetContainerStats(ctr.Names[0])), 40, 2, true, false)

			app.SetFocus(shellCommandInput)
		})
	}

	containerView.SetBorder(true).SetTitle("Containers")
	infoView.SetBorder(true).SetTitle("Info")
	shellView.SetBorder(true).SetTitle("Shell")
	statsView.SetBorder(true).SetTitle("Stats")

	shellOutputView.SetDynamicColors(true)
	shellOutputView.SetRegions(true)
	shellOutputView.SetWrap(true)
	shellOutputView.SetChangedFunc(func() {
		app.Draw()
	})

	shellCommandInput.SetFieldBackgroundColor(tcell.ColorBlack)
	shellCommandInput.SetLabel("$ ")
	shellCommandInput.SetFieldWidth(100)
	shellCommandInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			command := shellCommandInput.GetText()
			output := executeCommand(currentContainer, command)
			shellOutputView.SetText(shellOutputView.GetText(true) + "$ " + command + "\n" + output + "\n")
			shellCommandInput.SetText("")
			shellOutputView.ScrollToEnd()
			app.SetFocus(shellOutputView)
		}
	})

	shellView.SetDirection(tview.FlexRow)
	shellView.AddItem(shellOutputView, 0, 1, true)
	shellView.AddItem(shellCommandInput, 1, 1, true)

	commandInput := tview.NewInputField()
	commandInput.SetFieldBackgroundColor(tcell.ColorBlack)
	commandInput.SetFieldWidth(100)
	commandInput.SetDoneFunc(func (key tcell.Key) {
		if key == tcell.KeyEnter {
			//command := commandInput.GetText()
			commandInput.SetText("")
			layout.RemoveItem(commandInput)
			app.SetFocus(containerView)
		}
	})

	/*main.AddItem(tview.NewBox().SetBorder(true).SetTitle("Containers"), 0, 1, false)
	main.AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Info"), 0, 1, false).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Shell"), 0, 2, false).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Stats"), 0, 1, false), 0, 2, false)*/

	main.AddItem(containerView, 0, 1, false)
	main.AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(infoView, 0, 1, false).
		AddItem(shellView, 0, 1, false).
		AddItem(statsView, 0, 1, false), 0, 2, false)

	layout.AddItem(main, 0, 44, false)
	layout.SetDirection(tview.FlexRow)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == ':' {
			layout.AddItem(commandInput, 0, 1, true)
			app.SetFocus(commandInput)
		}
		return event
	})

	if err := app.SetRoot(layout, true).SetFocus(containerView).Run(); err != nil {
		panic(err)
	}
}

func GetContainerStats(container string) float64 {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	defer apiClient.Close()

	os.Setenv("DOCKER_API_VERSION", "1.45")

	stats, err := apiClient.ContainerStats(context.Background(), strings.Replace(container, "/", "", -1), false)
	if err != nil {
		panic(err)
	}

	defer stats.Body.Close()

	data, err := io.ReadAll(stats.Body)
	if err != nil {
		panic(err)
	}

	var statsJSON types.StatsJSON
	err = json.Unmarshal(data, &statsJSON)
	if err != nil {
		panic(err)
	}

	cpuDelta := float64(statsJSON.CPUStats.CPUUsage.TotalUsage) - float64(statsJSON.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(statsJSON.CPUStats.SystemUsage) - float64(statsJSON.PreCPUStats.SystemUsage)
	numberOfCores := float64(statsJSON.CPUStats.OnlineCPUs)

	cpuPercent := (cpuDelta / systemDelta) * numberOfCores * 100.0
	return cpuPercent
}

func executeCommand(container string, command string) string {
	cmd := exec.Command("docker", "exec", strings.Replace(container, "/", "", -1), command)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return out.String()
		//return err.Error()
	}
	return out.String()
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
			execDockerCompose(file.Name())
		}
	}
}

func execDockerCompose(composeFile string) {
	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		log.Fatalf("%s file not found", composeFile)
	}

	cmd := exec.Command("docker", "compose", "-f", composeFile, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
    cmd.Dir = "docker-compose"

	if err := cmd.Run(); err != nil {
		log.Fatalf("Error in executing docker compose: %s", err)
	}

	fmt.Println(composeFile, " successfully started")
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

func GetAllContainers() []types.Container {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	defer apiClient.Close()

	os.Setenv("DOCKER_API_VERSION", "1.45")
	
	containers, err := apiClient.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		panic(err)
	}

	return containers;
}

func mkdir(dirPath string) {
	err := os.MkdirAll(dirPath, 0755)

	if err != nil {
		fmt.Printf("Errore nella creazone della directory: %v\n", err)
	} else {
		fmt.Printf("Directory %s creata con successo.\n", dirPath)
	}
}


