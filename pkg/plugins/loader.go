package plugins

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"

	"gopkg.in/yaml.v2"
	"helm.sh/helm/v4/internal/plugins/runtimes/subprocesslegacy"
	"helm.sh/helm/v4/pkg/helmpath"
)

type pluginFSLoader struct {
	root os.Root
}

// What is a plugin?
// - files on disk in a directory
// - for "new" plugins, containing at least a plugin.yaml manifest and a signature
// - for legacy subprocess plugins, a plugin.yaml manifest is inferred (generated)

func isSubprocessLegacyPlugin(f fs.StatFS) bool {
	_, err := f.Stat("plugin.yaml")
	_, err := f.Stat("plugin.yaml")

}

func loadPluginDir(pluginsRoot os.Root, pluginDirName string) (pluginRaw, error) {

	pluginRoot, err := pluginsRoot.OpenRoot(pluginDirName)
	if err != nil {
		return pluginRaw{}, fmt.Errorf("failed to open plugin dir %q: %w", pluginDirName, err)
	}

	pfs := pluginFS{root: *pluginRoot}

	// Read the plugin manifest
	manifest, err := pfs.readManifestFile()
	if err != nil {
		return pluginRaw{}, err
	}

	return pluginRaw{
		root:     *pluginRoot,
		manifest: manifest,
	}, nil
}

type pluginRaw struct {
	root     os.Root
	manifest PluginManifest
}

type pluginsFSLoaderCallback func(*pluginRaw) error

func (p *pluginFSLoader) LoadPlugins(cb pluginsFSLoaderCallback) error {
	// enumerate plugins on disk
	//   for each plugin dir/manifest:
	// 	  - read plugin manifest /or infer/generate one for legacy plugins
	//    - determine runtime then call runtime with plugin location
	//    - runtime then does what it needs for

	pluginDirEntries, err := p.root.FS().(fs.ReadDirFS).ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read plugins directory: %w", err)
	}

	for _, pluginDirEntry := range pluginDirEntries {
		if !pluginDirEntry.IsDir() {
			continue // skip non-directory entries
		}

		pluginDirName := pluginDirEntry.Name()

		pluginRaw, err := loadPluginDir(p.root, pluginDirName)
		if err != nil {
			return err
		}

		err := cb(&pluginRaw)
		if err != nil {
			return err
		}
	}

	return nil
}

type pluginFS struct {
	root os.Root
}

const pluginManifestFile = "plugin.yaml"

func (p *pluginFS) openPluginManifest() (fs.File, error) {
	f, err := p.dir.Open(pluginManifestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin manifest file (%s): %w", pluginManifestFile, err)
	}

	return f, nil
}

const pluginsAPIVersion = "plugins.helm.sh/v1alpha1"

func peekManifestYAMLAPIVersion(r io.Reader) (string, error) {
	type apiVersion struct {
		APIVersion string `yaml:"APIVersion"`
	}

	var v apiVersion
	d := yaml.NewDecoder(r)
	if err := d.Decode(v); err != nil {
		return "", fmt.Errorf("failed to decode plugin manifest file (%s) yaml: %w", pluginManifestFile, err)
	}

	return v.APIVersion, nil
}

func decodePluginManifestYAML(r io.Reader) (PluginManifest, error) {

	var v PluginManifest
	d := yaml.NewDecoder(r)
	if err := d.Decode(v); err != nil {
		return PluginManifest{}, fmt.Errorf("failed to decode plugin manifest file (%s) yaml: %w", pluginManifestFile, err)
	}

	return v, nil
}

func (p *pluginFS) readManifestFile() (PluginManifest, error) {
	f, err := p.openPluginManifest()
	if err != nil {
		return PluginManifest{}, err
	}
	defer f.Close()

	var b []byte
	if _, err := f.Read(b); err != nil {
		return PluginManifest{}, fmt.Errorf("failed to read plugin manifest file (%s): %w", pluginManifestFile, err)
	}

	apiVersion, err := peekManifestYAMLAPIVersion(bytes.NewReader(b))
	if err != nil {
		return PluginManifest{}, fmt.Errorf("failed to peek plugin manifest file (%s) yaml: %w", pluginManifestFile, err)
	}

	if apiVersion == "" { // subprocess legacy plugin
		lp, err := subprocesslegacy.LoadDir(p.root.Name())
		if err != nil {
			return PluginManifest{}, fmt.Errorf("failed to load subprocess legacy plugin from dir %q: %w", p.root.Name(), err)
		}

		if len(lp.Metadata.Downloaders) > 0 {
			convertDownloaderSchemes := func(downloasders []subprocesslegacy.Downloaders) []string {
				results := []string{}
				for _, d := range downloasders {
					results = append(results, d.Protocols...)
				}

				return results
			}
			return PluginManifest{
				APIVersion:   pluginsAPIVersion,
				Name:         lp.Metadata.Name,
				Version:      lp.Metadata.Version,
				Description:  lp.Metadata.Description,
				Type:         "downloader/v1",
				RuntimeClass: "subprocesslegacy",
				Config: map[string]any{
					"downloader_schemes": convertDownloaderSchemes(lp.Metadata.Downloaders),
				},
			}, nil
		}
	}

	return decodePluginManifestYAML(bytes.NewReader(b))
}

type pluginManager struct {
	runtimes map[string]pluginRuntime
	catalog  pluginCatalogImpl
}

func (p *pluginManager) init() {
	// Register runtimes
	p.runtimes["subprocesslegacy"] = subprocesslegacy.NewRuntime()

	// Load plugins
	pluginsRoot, err := os.OpenRoot(helmpath.DataPath("plugins"))
	if err != nil {
		panic(fmt.Errorf("failed to open plugins directory: %w", err))
	}
	defer pluginsRoot.Close()

	pluginLoader := pluginFSLoader{*pluginsRoot}
	pluginLoader.LoadPlugins(func(pr *pluginRaw) error {
		// Register each plugin

		r, ok := p.runtimes[pr.manifest.RuntimeClass]
		if !ok {
			return fmt.Errorf("unknown runtime class %q for plugin %q", pr.manifest.RuntimeClass, pr.root.Name())
		}

		if err := r.ValidatePlugin(pr.root); err != nil {
			return fmt.Errorf("failed to validate plugin %q: %w", pr.root.Name(), err)
		}

		p.catalog.registerPlugin(pr, r)

		return nil
	})

	//// Initialize catalog
	//for rn, r := range p.runtimes {
	//	r.All()

	//}
}

func (p *pluginManager) Catalog() PluginCatalog {
	return &p.catalog
}
