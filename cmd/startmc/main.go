package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type config struct {
	javaPath   string
	port       string
	serverPath string
	screenName string
}

type Response struct {
	Success      bool   `json:"success,omitempty"`
	IsRunning    bool   `json:"is_running"`
	Message      string `json:"message"`
	ErrorMessage string `json:"error,omitempty"`
}

var cfg = config{
	javaPath:   getEnv("STARTMC_JAVA_PATH", "java"),
	port:       getEnv("STARTMC_PORT", "80"),
	serverPath: getRequiredEnv("MC_SERVER_PATH"),
	screenName: getEnv("MC_SCREEN_NAME", "mcserver"),
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handleToggle)
	mux.HandleFunc("GET /status", handleStatus)

	addr := ":" + cfg.port
	log.Printf("Server starting on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	running, err := isMCRunning()
	if err != nil {
		sendJSONResponse(w, Response{
			IsRunning:    running,
			Message:      "Failed to check server status",
			ErrorMessage: err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	message := "Minecraft server is stopped"
	if running {
		message = "Minecraft server is running"
	}
	sendJSONResponse(w, Response{
		IsRunning: running,
		Message:   message,
	}, http.StatusOK)
}

func handleToggle(w http.ResponseWriter, r *http.Request) {
	running, err := isMCRunning()
	if err != nil {
		sendJSONResponse(w, Response{
			Success:      false,
			IsRunning:    running,
			Message:      "Failed to check server status",
			ErrorMessage: err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	if running {
		if err := stopMC(); err != nil {
			sendJSONResponse(w, Response{
				Success:      false,
				IsRunning:    true,
				Message:      "Failed to stop server",
				ErrorMessage: err.Error(),
			}, http.StatusInternalServerError)
			return
		}
		log.Printf("Stopped server")
		sendJSONResponse(w, Response{
			Success:   true,
			IsRunning: false,
			Message:   "Minecraft server stopped successfully",
		}, http.StatusOK)
	} else {
		if err := startMC(); err != nil {
			sendJSONResponse(w, Response{
				Success:      false,
				IsRunning:    false,
				Message:      "Failed to start server",
				ErrorMessage: err.Error(),
			}, http.StatusInternalServerError)
			return
		}
		log.Printf("Started server")
		sendJSONResponse(w, Response{
			Success:   true,
			IsRunning: true,
			Message:   "Minecraft server started successfully",
		}, http.StatusOK)
	}
}

func isMCRunning() (bool, error) {
	cmd := exec.Command("screen", "-ls")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// screen -ls returns error if no sessions exist, which is fine
		if strings.Contains(string(output), "No Sockets found") {
			return false, nil
		}
		return false, fmt.Errorf("failed to list screen sessions: %v", err)
	}

	return strings.Contains(string(output), cfg.screenName), nil
}

func startMC() error {
	cmd := exec.Command("screen", "-dmS", cfg.screenName, cfg.javaPath,
		"-Xms1G", "-Xmx3G", "-jar", cfg.serverPath+"server.jar", "nogui")
	cmd.Dir = cfg.serverPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start Minecraft server: %v", err)
	}
	return nil
}

func stopMC() error {
	cmd := exec.Command("screen", "-S", cfg.screenName, "-X", "stuff", "\nstop\n")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop Minecraft server: %v", err)
	}
	return nil
}

func sendJSONResponse(w http.ResponseWriter, response any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func getRequiredEnv(key string) string {
	val, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("missing required env var %s", key)
	}

	return val
}

func getEnv(key, fallback string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}

	return fallback
}
