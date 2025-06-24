package getter

import (
	"bytes"
	"context"

	"helm.sh/helm/v4/pkg/cli"
	"helm.sh/helm/v4/pkg/plugins"
)

func collectDownloaderPlugins(settings *cli.EnvSettings) Providers {

	pluginDescriptor := plugins.PluginDescriptor{
		Type:    "downloaders",
		Version: "v1",
	}

	plgs, err := settings.PluginCatalog.CollectPlugins(pluginDescriptor)
	if err != nil {
	}

	pluginConstructorBuilder := func(plg plugins.Plugin) Constructor {
		return func(option ...Option) (Getter, error) {

			return &getterPlugin{
				options: append([]Option{}, option...),
				plg:     plg,
			}, nil
		}
	}

	results := make([]Provider, len(plgs))

	for _, plg := range plgs {

		downloaderSchemes, ok := (plg.Manifest().Config["downloader_schemes"]).([]string)
		if !ok {
		}

		results = append(results, Provider{
			Schemes: downloaderSchemes,
			New:     pluginConstructorBuilder(plg),
		})
	}

	return results
}

type getterPlugin struct {
	options []Option
	plg     plugins.Plugin
}

func (g *getterPlugin) Get(url string, options ...Option) (*bytes.Buffer, error) {

	// TODO: can we generate these? (plugin input/outputs)
	type getterInputV1 struct {
		URL     string   `json:"url"`
		Options []Option `json:"options"`
	}

	// TODO: can we generate these? (plugin input/outputs)
	type getterOutputV1 struct {
		Data *bytes.Buffer `json:"data"`
	}

	o := make([]Option, len(g.options)+len(options))
	input := getterInputV1{
		URL:     url,
		Options: append(append(o, g.options...), options...),
	}
	var output getterOutputV1
	err := g.plg.Invoke(context.Background(), input, output)

	return output.Data, err
}
