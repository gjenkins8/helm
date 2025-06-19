package plugins

var (
	PluginCatalogNoMatchingPlugin = errors.New("no matching plugin found")

	PluginCatalogMultipleMatchingPlugins = errors.New("multiple matching plugins found")
)

type PluginCatalog interface {
	// CollectPlugin returns a single plugin matching the given plugin descriptor
	//
	// Errors:
	// - PluginCatalogNoMatchingPlugin is returned if no matching plugin is found
	// - PluginCatalogMultipleMatchingPlugins is returned if multiple matching plugins are found
	CollectPlugin(descriptor plugins.PluginDescriptor) (plugins.PluginInstance, error)

	// CollectPlugins finds all the plugins matching the given plguin descriptor
	//
	// Errors:
	// - PluginCatalogNoMatchingPlugin is returned if no matching plugin is found
	CollectPlugins(descriptor plugins.PluginDescriptor) ([]plugins.PluginInstance, error)
}
