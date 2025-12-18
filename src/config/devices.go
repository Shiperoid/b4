package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/daniellavrushin/b4/log"
)

type DeviceAliases struct {
	path    string
	aliases map[string]string
	mu      sync.RWMutex
}

func NewDeviceAliases(configPath string) *DeviceAliases {
	dir := filepath.Dir(configPath)
	path := filepath.Join(dir, "mac_aliases.json")

	da := &DeviceAliases{
		path:    path,
		aliases: make(map[string]string),
	}
	da.load()
	return da
}

func (da *DeviceAliases) load() {
	data, err := os.ReadFile(da.path)
	if err != nil {
		return
	}

	da.mu.Lock()
	defer da.mu.Unlock()

	if err := json.Unmarshal(data, &da.aliases); err != nil {
		log.Errorf("Failed to parse mac_aliases.json: %v", err)
	}
}

func (da *DeviceAliases) save() error {
	da.mu.RLock()
	data, err := json.MarshalIndent(da.aliases, "", "  ")
	da.mu.RUnlock()

	if err != nil {
		return err
	}
	return os.WriteFile(da.path, data, 0644)
}

func (da *DeviceAliases) Get(mac string) (string, bool) {
	da.mu.RLock()
	defer da.mu.RUnlock()
	alias, ok := da.aliases[mac]
	return alias, ok
}

func (da *DeviceAliases) Set(mac, name string) error {
	da.mu.Lock()
	da.aliases[mac] = name
	da.mu.Unlock()
	return da.save()
}

func (da *DeviceAliases) Delete(mac string) error {
	da.mu.Lock()
	delete(da.aliases, mac)
	da.mu.Unlock()
	return da.save()
}

func (da *DeviceAliases) GetAll() map[string]string {
	da.mu.RLock()
	defer da.mu.RUnlock()
	copy := make(map[string]string, len(da.aliases))
	for k, v := range da.aliases {
		copy[k] = v
	}
	return copy
}
