package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/daniellavrushin/b4/log"
)

func (api *API) RegisterSystemApi() {
	api.mux.HandleFunc("/api/system/restart", api.handleRestart)
	api.mux.HandleFunc("/api/system/info", api.handleSystemInfo)
	api.mux.HandleFunc("/api/version", api.handleVersion)
	api.mux.HandleFunc("/api/system/update", api.handleUpdate)
}

// detectServiceManager determines which service manager is managing B4
func detectServiceManager() string {
	// Check for systemd
	if _, err := os.Stat("/etc/systemd/system/b4.service"); err == nil {
		if _, err := exec.LookPath("systemctl"); err == nil {
			return "systemd"
		}
	}

	// Check for Entware/OpenWRT init script
	if _, err := os.Stat("/opt/etc/init.d/S99b4"); err == nil {
		return "entware"
	}

	// Check for standard init script
	if _, err := os.Stat("/etc/init.d/b4"); err == nil {
		return "init"
	}

	// Check if running as a standalone process (no service manager)
	return "standalone"
}

func (api *API) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	serviceManager := detectServiceManager()
	canRestart := serviceManager != "standalone"

	info := SystemInfo{
		ServiceManager: serviceManager,
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
		CanRestart:     canRestart,
	}

	setJsonHeader(w)
	json.NewEncoder(w).Encode(info)
}

func (api *API) handleRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	serviceManager := detectServiceManager()
	log.Infof("Restart requested via web UI (service manager: %s)", serviceManager)

	var response RestartResponse
	response.ServiceManager = serviceManager

	switch serviceManager {
	case "systemd":
		response.Success = true
		response.Message = "Restart initiated via systemd"
		response.RestartCommand = "systemctl restart b4"

	case "entware":
		response.Success = true
		response.Message = "Restart initiated via Entware init script"
		response.RestartCommand = "/opt/etc/init.d/S99b4 restart"

	case "init":
		response.Success = true
		response.Message = "Restart initiated via init script"
		response.RestartCommand = "/etc/init.d/b4 restart"

	case "standalone":
		response.Success = false
		response.Message = "Cannot restart: B4 is not running as a service. Please restart manually."
		setJsonHeader(w)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return

	default:
		response.Success = false
		response.Message = "Unknown service manager"
		setJsonHeader(w)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Send response immediately before triggering restart
	setJsonHeader(w)
	json.NewEncoder(w).Encode(response)

	// Flush the response to ensure it's sent
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Trigger restart in a goroutine with a small delay
	// This allows the HTTP response to be sent before the service stops
	go func() {
		time.Sleep(500 * time.Millisecond)
		log.Infof("Executing restart command: %s", response.RestartCommand)

		var cmd *exec.Cmd
		switch serviceManager {
		case "systemd":
			cmd = exec.Command("systemctl", "restart", "b4")
		case "entware":
			cmd = exec.Command("/opt/etc/init.d/S99b4", "restart")
		case "init":
			cmd = exec.Command("/etc/init.d/b4", "restart")
		}

		if cmd != nil {
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Errorf("Restart command failed: %v\nOutput: %s", err, string(output))
			} else {
				log.Infof("Restart command executed successfully")
			}
		}
	}()
}

func (api *API) handleVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	versionInfo := VersionInfo{
		Version:   Version,
		Commit:    Commit,
		BuildDate: Date,
	}
	setJsonHeader(w)
	enc := json.NewEncoder(w)
	_ = enc.Encode(versionInfo)
}

func (api *API) handleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	serviceManager := detectServiceManager()
	log.Infof("Update requested via web UI (service manager: %s, version: %s)", serviceManager, req.Version)

	var response UpdateResponse
	response.ServiceManager = serviceManager

	// Check if we can perform updates
	if serviceManager == "standalone" {
		response.Success = false
		response.Message = "Cannot update: B4 is not running as a service. Please update manually."
		setJsonHeader(w)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Prepare update command based on service manager
	var updateCmd string
	switch serviceManager {
	case "entware":
		log.Infof("Preparing update command for Entware service manager")
		// Use the built-in update command in the init script
		updateCmd = "/opt/etc/init.d/S99b4 update"
		response.UpdateCommand = updateCmd

	case "systemd", "init":
		log.Infof("Preparing update command for service manager: %s", serviceManager)
		// For systemd and standard init, we'll download and run the installer
		// The installer will handle stopping/starting the service
		updateCmd = "wget -O /tmp/b4install.sh https://raw.githubusercontent.com/DanielLavrushin/b4/main/install.sh && chmod +x /tmp/b4install.sh && /tmp/b4install.sh -q"
		response.UpdateCommand = updateCmd

	default:
		response.Success = false
		response.Message = "Unknown service manager"
		setJsonHeader(w)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response.Success = true
	response.Message = "Update initiated. The service will restart automatically."

	// Send response immediately before triggering update
	setJsonHeader(w)
	json.NewEncoder(w).Encode(response)

	// Flush the response to ensure it's sent
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Trigger update in a goroutine with a small delay
	// This allows the HTTP response to be sent before the service stops
	go func() {
		time.Sleep(500 * time.Millisecond)
		log.Infof("Executing update command: %s", updateCmd)

		var cmd *exec.Cmd
		switch serviceManager {
		case "entware":
			cmd = exec.Command("/opt/etc/init.d/S99b4", "update")
		case "systemd", "init":
			// Use sh to execute the compound command
			cmd = exec.Command("sh", "-c", updateCmd)
		}

		if cmd != nil {
			// Set up command to run independently
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			// Start the command but don't wait for it
			// The update process will kill this process anyway
			if err := cmd.Start(); err != nil {
				log.Errorf("Update command failed to start: %v", err)
			} else {
				log.Infof("Update command started successfully (PID: %d)", cmd.Process.Pid)
				// Don't wait for the command - let it run independently
			}
		}
	}()
}
