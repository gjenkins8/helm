package plugins

import (
	"errors"
	"fmt"
)

type HostFunction struct {
	Name    string
	Version string
}

type PluginStore interface {
}

type PluginCatalog interface {
	Lookup(PluginDescriptor) ([]PluginManifest, error)
}

type PluginDescriptor struct {
	Class   string
	Version string
}

func (d *PluginDescriptor) String() string {
	return fmt.Sprintf("[%q %q]", d.Class, d.Version)
}

type PluginLoader interface {
	LoadPlugin(PluginManifest) (Plugin, error)
}

type PluginManager struct {
	catalog PluginCatalog
	loader  PluginLoader
}

// RetrievePlugins retrives zero or more plugins for the given plugin descriptors
func (p *PluginManager) RetrievePlugins(descriptors ...PluginDescriptor) ([]Plugin, error) {

	var results []Plugin
	var errs []error
	for _, descriptor := range descriptors {
		pluginManifests, err := p.catalog.Lookup(descriptor)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to lookup plugins: %s: %w", descriptor.String(), err))

			continue
		}

		if len(pluginManifests) == 0 {
			continue
		}

		for _, pluginManifest := range pluginManifests {
			plugin, err := p.loader.LoadPlugin(pluginManifest)
			if err != nil {

				errs = append(errs, fmt.Errorf("failed to load plugin: %s %w", pluginManifest.String(), err))
				continue
			}

			results = append(results, plugin)
		}

	}

	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("failed to retrieve plugins: %w", err)
	}

	return results, nil
}

// RetrievePlugin retrives one plugins for the given plugin descriptor, returning an error if exactly one plugin not found
func (p *PluginManager) RetrievePlugin(descriptor PluginDescriptor) (Plugin, error) {

	pluginManifests, err := p.catalog.Lookup(descriptor)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup plugins: %s: %w", descriptor.String(), err)
	}

	if len(pluginManifests) > 1 {
		return nil, fmt.Errorf("failed to lookup plugins: %s: %w", descriptor.String(), err)
	}
	if len(pluginManifests) == 0 {
		return nil, fmt.Errorf("failed to lookup plugins: %s: %w", descriptor.String(), err)
	}

	pluginManifest := pluginManifests[0]
	plugin, err := p.loader.LoadPlugin(pluginManifest)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin: %s %w", pluginManifest.String(), err)
	}

	return plugin, nil
}
