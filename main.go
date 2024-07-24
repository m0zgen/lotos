package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// Config struct for YAML configuration
type Config struct {
	Port        int    `yaml:"port"`
	LogFilePath string `yaml:"logFilePath"`
	ShowLogs    bool   `yaml:"showLogs"`
}

var config Config

// Clients map to store connected clients
var clients = make(map[*websocket.Conn]bool)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// readConfig- Read configuration from YAML file
func readConfig(filename string) (Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// handleConnections - Handle WebSocket connections
func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to websocket: %v", err)
		return
	}
	defer ws.Close()

	clients[ws] = true
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			delete(clients, ws)
			break
		}
	}
}

// handleFileChanges - Handle file changes
func handleFileChanges(logFile string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to create file watcher: %v", err)
	}
	defer watcher.Close()

	err = watcher.Add(logFile)
	if err != nil {
		log.Fatalf("Failed to add file to watcher: %v", err)
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				data, err := ioutil.ReadFile(logFile)
				if err != nil {
					log.Printf("Error reading file: %v", err)
					continue
				}
				message := string(data)
				if config.ShowLogs {
					log.Printf("Sending message: %s", message)
				}
				for client := range clients {
					err := client.WriteMessage(websocket.TextMessage, []byte(message))
					if err != nil {
						log.Printf("Failed to send message: %v", err)
						client.Close()
						delete(clients, client)
					}
				}
			}
		case err := <-watcher.Errors:
			log.Printf("Watcher error: %v", err)
		}
	}
}

// runWebSocketServer - Run WebSocket server
func runWebSocketServer() {

	// Setup HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "WebSocket server is running.")
	})

	http.HandleFunc("/ws", handleConnections)

	go handleFileChanges(config.LogFilePath)

	addr := fmt.Sprintf(":%d", config.Port)
	log.Printf("Server is running on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func main() {
	var err error

	// Check if the config file is provided from the command line
	if len(os.Args) < 2 {
		log.Println("Usage: go run main.go <config.yml>")
		return
	}

	configFile := os.Args[1]

	config, err = readConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		os.Exit(1)
	}

	// Run WebSocket server with handlers
	runWebSocketServer()

}
