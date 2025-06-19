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

type PluginManifest struct {
	// Kind of the plugin: 'downloader', 'postrenderer', etc
	// Describing the situation the plugin is expected to be invoked, and the correspondng message types used to invoke
	Kind string

	// Version of the plugin kind
	Version string

	// Runtime used to execute the plugin
	RuntimeClass string

	// Additional config associated with the plugin kind: e.g. downloader URI schemes
	// (Config is intepreted by the plugin invoker)
	Config map[string]any
}

//type PluginClass interface {
//	CreateInstance() PluginInstance
//}

type PluginInstance interface {
	Manifest() PluginManifest
	Invoke(ctx context.Context, input any, output any) error
}
