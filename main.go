package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// http://localhost:3001
// 192.168.0.189
var testLink string = "http://192.168.0.189:3001"
var mainLink string = testLink
var stationName string

type ResponseBody struct {
	Ok int `json:"ok"`
}

func main() {
	fmt.Println("Enter the name of the station:")
	fmt.Scanln(&stationName)

	executablePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v", err)
	}

	baseDirectory := filepath.Dir(executablePath)
	audioDirectory := filepath.Join(baseDirectory, "audio")
	fileNames, err := listFilesInDirectory(audioDirectory)
	if err != nil {
		log.Fatalf("Error listing files in directory: %v", err)
	}

	responseJSON, err := json.Marshal(fileNames)
	if err != nil {
		log.Fatalf("Error serializing file list to JSON: %v", err)
	}

	requestBody := bytes.NewBuffer(responseJSON)

	ipAddress, err := getIPAddress()
	if err != nil {
		log.Fatalf("Error getting IP address: %v", err)
	}

	response, err := http.Post(mainLink+"/connect/"+stationName+"?ip="+ipAddress, "application/json", requestBody)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	var responseBody ResponseBody
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	if responseBody.Ok == 200 {
		fmt.Println("Server response: OK (status 200)")
		go startHTTPServer() // Uruchomienie serwera HTTP w osobnej gorutynie
	} else {
		fmt.Println("Server response: Invalid status")
	}

	// Poczekaj, aby uniknąć zakończenia działania programu
	select {}
}

func getIPAddress() (string, error) {
	response, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var data map[string]string
	if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
		return "", err
	}

	ipAddress, ok := data["ip"]
	if !ok {
		return "", fmt.Errorf("IP address not found in response")
	}

	return ipAddress, nil
}

func startHTTPServer() {
	fmt.Println("Starting HTTP server...")
	http.HandleFunc("/", handleAudioRequest)
	if err := http.ListenAndServe(":3002", nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}

func handleAudioRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request from server")
	fileName := filepath.Base(r.URL.Path)

	audioDirectory := "./audio"
	filePath := filepath.Join(audioDirectory, fileName)
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "audio/mpeg")

	_, err = io.Copy(w, file)
	if err != nil {
		log.Printf("Error copying file content to HTTP response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func listFilesInDirectory(directoryPath string) ([]string, error) {
	var fileNames []string

	files, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileNames = append(fileNames, file.Name())
	}

	return fileNames, nil
}
