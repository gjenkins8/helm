package plugins

import (
	"context"
	"fmt"
)

type PluginInstance interface {
	Plugin() Plugin
	Invoke(ctx context.Context, input any, output any) error
}

type PluginManifest struct {
	Kind          string
	APIVersion    string
	Class         string
	PluginName    string
	Version       string
	Hooks         []string
	HostFunctions []string
	Config        map[string]string
}

func (m *PluginManifest) String() string {
	return fmt.Sprintf("[%q %q %q]", m.PluginName, m.Class, m.Version)
}

type Plugin interface {
	Manifest() PluginManifest
	CreateInstance() (PluginInstance, error)
}

type HookDescriptor struct {
	HookName   string
	APIVersion string
}

type Descriptor struct {
	PluginName string
}

type Catalog interface {
	GetPlugin(HookDescriptor) (Plugin, error)
	GetPlugins(HookDescriptor) ([]Plugin, error)
}

type Config struct {
}
