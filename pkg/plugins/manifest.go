package plugins

// plugin.yaml definition
type PluginManifest struct {
	// APIVersion of the plugin manifest document
	// Currently: 'plugins.helm.sh/v1alpha1'
	APIVersion string

	// Author defined name, version and description of the plugin
	Name        string
	Version     string
	Description string

	// Type of the plugin: 'downloader/v1', 'postrenderer/v1', etc
	// Describing the situation the plugin is expected to be invoked, and the correspondng message type/version used to invoke
	Type string

	// Runtime used to execute the plugin
	// subprocesslegacy, extism/v1, etc
	RuntimeClass string

	// Additional config associated with the plugin kind: e.g. downloader URI schemes
	// (Config is intepreted by the plugin invoker)
	Config map[string]any
}
