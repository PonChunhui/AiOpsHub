package agent

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aiops/AiOpsHub/backend/internal/model"
)

type ToolFactoryFunc func(tool *model.Tool, overrideConfig string) (Tool, error)

var (
	toolFactories = make(map[string]ToolFactoryFunc)
	factoryMutex  sync.RWMutex
)

func RegisterToolFactory(name string, factory ToolFactoryFunc) {
	factoryMutex.Lock()
	toolFactories[name] = factory
	factoryMutex.Unlock()
}

func GetToolFactory(name string) (ToolFactoryFunc, bool) {
	factoryMutex.RLock()
	factory, exists := toolFactories[name]
	factoryMutex.RUnlock()
	return factory, exists
}

func CreateTool(tool *model.Tool, overrideConfig string) (Tool, error) {
	factory, exists := GetToolFactory(tool.Name)
	if !exists {
		return nil, fmt.Errorf("tool factory not found: %s", tool.Name)
	}

	return factory(tool, overrideConfig)
}

func ParseConfig(defaultConfig, override string) map[string]interface{} {
	config := make(map[string]interface{})

	if defaultConfig != "" {
		json.Unmarshal([]byte(defaultConfig), &config)
	}

	if override != "" {
		overrideMap := make(map[string]interface{})
		json.Unmarshal([]byte(override), &overrideMap)

		for k, v := range overrideMap {
			config[k] = v
		}
	}

	return config
}
