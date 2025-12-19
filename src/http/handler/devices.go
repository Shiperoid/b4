package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	instance *VendorLookup
	once     sync.Once
)

type VendorInfo struct {
	Company string `json:"company"`
	Country string `json:"country"`
}

type VendorLookup struct {
	cache  map[string]VendorInfo
	mu     sync.RWMutex
	client *http.Client
}

type DeviceInfo struct {
	MAC        string `json:"mac"`
	IP         string `json:"ip"`
	Hostname   string `json:"hostname"`
	Vendor     string `json:"vendor"`
	Country    string `json:"country"`
	DeviceType string `json:"device_type"`
	IsPrivate  bool   `json:"is_private"`
	Alias      string `json:"alias,omitempty"`
}
type DevicesResponse struct {
	Available bool         `json:"available"`
	Source    string       `json:"source,omitempty"`
	Devices   []DeviceInfo `json:"devices"`
}

func (api *API) RegisterDevicesApi() {
	api.mux.HandleFunc("/api/devices", api.handleDevices)
	api.mux.HandleFunc("/api/devices/{mac}/vendor", api.handleDeviceVendor)
	api.mux.HandleFunc("/api/devices/{mac}/alias", api.handleDeviceAlias)
}

func (api *API) handleDeviceAlias(w http.ResponseWriter, r *http.Request) {
	mac := r.PathValue("mac")
	if mac == "" {
		http.Error(w, "MAC address required", http.StatusBadRequest)
		return
	}

	mac = normalizeMAC(mac)
	if len(mac) == 12 {
		mac = fmt.Sprintf("%s:%s:%s:%s:%s:%s", mac[0:2], mac[2:4], mac[4:6], mac[6:8], mac[8:10], mac[10:12])
	}

	switch r.Method {
	case http.MethodGet:
		alias, ok := api.deviceAliases.Get(mac)
		setJsonHeader(w)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mac":       mac,
			"alias":     alias,
			"has_alias": ok,
		})

	case http.MethodPut, http.MethodPost:
		var req struct {
			Alias string `json:"alias"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if req.Alias == "" {
			http.Error(w, "Alias cannot be empty", http.StatusBadRequest)
			return
		}

		if err := api.deviceAliases.Set(mac, req.Alias); err != nil {
			http.Error(w, "Failed to save alias", http.StatusInternalServerError)
			return
		}

		setJsonHeader(w)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"mac":     mac,
			"alias":   req.Alias,
		})

	case http.MethodDelete:
		if err := api.deviceAliases.Delete(mac); err != nil {
			http.Error(w, "Failed to delete alias", http.StatusInternalServerError)
			return
		}

		setJsonHeader(w)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"mac":     mac,
		})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (api *API) handleDeviceVendor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	mac := r.PathValue("mac")

	if mac == "" {
		http.Error(w, "MAC address required", http.StatusBadRequest)
		return
	}

	vendorLookup := getVendorLookup()
	info := vendorLookup.Lookup(mac)

	setJsonHeader(w)
	json.NewEncoder(w).Encode(info)
}

func (api *API) handleDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	setJsonHeader(w)

	if globalPool == nil || globalPool.Dhcp == nil || !globalPool.Dhcp.IsAvailable() {
		json.NewEncoder(w).Encode(DevicesResponse{
			Available: false,
			Devices:   []DeviceInfo{},
		})
		return
	}

	sourceName, _ := globalPool.Dhcp.SourceInfo()
	mappings := globalPool.Dhcp.GetAllMappings()

	vendorLookup := getVendorLookup()
	devices := make([]DeviceInfo, 0, len(mappings))

	for ip, macAddr := range mappings {
		var vendor, country string
		if isPrivateMAC(macAddr) {
			vendor = "Private"
		} else {
			info := vendorLookup.Lookup(macAddr)
			vendor = info.Company
			country = info.Country
		}

		alias, _ := api.deviceAliases.Get(macAddr)

		devices = append(devices, DeviceInfo{
			MAC:      macAddr,
			IP:       ip,
			Hostname: "",
			Vendor:   vendor,
			Alias:    alias,
			Country:  country,
		})
	}

	json.NewEncoder(w).Encode(DevicesResponse{
		Available: true,
		Source:    sourceName,
		Devices:   devices,
	})
}
func getVendorLookup() *VendorLookup {
	once.Do(func() {
		instance = &VendorLookup{
			cache: make(map[string]VendorInfo),
			client: &http.Client{
				Timeout: 3 * time.Second,
			},
		}
	})
	return instance
}

func normalizeMAC(mac string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(mac, ":", ""), "-", ""))
}

func (v *VendorLookup) Lookup(mac string) VendorInfo {
	normalized := normalizeMAC(mac)
	if len(normalized) < 6 {
		return VendorInfo{}
	}
	oui := normalized[:6]

	v.mu.RLock()
	if info, ok := v.cache[oui]; ok {
		v.mu.RUnlock()
		return info
	}
	v.mu.RUnlock()

	resp, err := v.client.Get(fmt.Sprintf("https://www.macvendorlookup.com/api/v2/%s/pipe", oui))
	if err != nil {
		return VendorInfo{}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		v.mu.Lock()
		v.cache[oui] = VendorInfo{}
		v.mu.Unlock()
		return VendorInfo{}
	}

	if resp.StatusCode != http.StatusOK {
		return VendorInfo{}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return VendorInfo{}
	}

	// Format: startHex|endHex|startDec|endDec|company|addr1|addr2|addr3|country|type
	parts := strings.Split(string(body), "|")
	if len(parts) < 9 {
		return VendorInfo{}
	}

	info := VendorInfo{
		Company: parts[4],
		Country: parts[8],
	}

	v.mu.Lock()
	v.cache[oui] = info
	v.mu.Unlock()

	return info
}

func (v *VendorLookup) BulkLookup(macs []string) map[string]VendorInfo {
	result := make(map[string]VendorInfo)

	for i, mac := range macs {
		result[mac] = v.Lookup(mac)
		if i < len(macs)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}
	return result
}

func isPrivateMAC(mac string) bool {
	normalized := normalizeMAC(mac)
	if len(normalized) < 2 {
		return false
	}
	secondChar := normalized[1]
	return secondChar == '2' || secondChar == '6' || secondChar == 'A' || secondChar == 'E'
}
