package plugins

import "context"

//type PluginManager interface {
//	RegisterPlugin(manifest plugins.PluginManifest, loader PluginLoader) error
//}
//
//
//type PluginLoader interface {
//	Load(location string, manifest plugins.PluginManifest) (plugins.PluginInstance, error)
//}
//
//type actualPluginManager struct {
//	loaders map[string] PluginLoader
//	plugins []Plugin
//}
//
//func (a *actualPluginManager) LoadPlugins() {
//	for pluginManifest in os.Find(...) {
//		loader := a.loaders[pluginManfest.Kind]
//		a.plugins = append(a.plugins, loader.Load(pluginManifest)
//	}
//}
//
//type ExtismPluginRunner type {
//}
//
//type ExtismPlugin type {
//}
//
//func (e *ExtismPlugin) Invoke(ctx context.Context, input any, output any) error {
//}

type PluginDescriptor struct {
	Type    string
	Version string
}

//type PluginClass interface {
//	CreateInstance() PluginInstance
//}

type Plugin interface {
	Manifest() PluginManifest
	Invoke(ctx context.Context, input any, output any) error
}
