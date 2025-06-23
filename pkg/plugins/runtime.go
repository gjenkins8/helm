package plugins

type pluginCatalogImpl struct{}

var _ PluginCatalog = (*pluginCatalogImpl)(nil)

func (p *pluginCatalogImpl) CollectPlugin(descriptor PluginDescriptor) (PluginInstance, error) {
	return nil, nil
}

func (p *pluginCatalogImpl) CollectPlugins(descriptor PluginDescriptor) ([]PluginInstance, error) {
	return nil, nil
}

type pluginRuntime interface {
}
