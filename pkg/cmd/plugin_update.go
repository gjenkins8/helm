/*
Copyright The Helm Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/spf13/cobra"

	plugin "helm.sh/helm/v4/pkg/legacyplugin"
	"helm.sh/helm/v4/pkg/legacyplugin/installer"
)

type pluginUpdateOptions struct {
	names []string
}

func newPluginUpdateCmd(out io.Writer) *cobra.Command {
	o := &pluginUpdateOptions{}

	cmd := &cobra.Command{
		Use:     "update <plugin>...",
		Aliases: []string{"up"},
		Short:   "update one or more Helm plugins",
		ValidArgsFunction: func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return compListPlugins(toComplete, args), cobra.ShellCompDirectiveNoFileComp
		},
		PreRunE: func(_ *cobra.Command, args []string) error {
			return o.complete(args)
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return o.run(out)
		},
	}
	return cmd
}

func (o *pluginUpdateOptions) complete(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("please provide plugin name to update")
	}
	o.names = args
	return nil
}

func (o *pluginUpdateOptions) run(out io.Writer) error {
<<<<<<< HEAD
	installer.Debug = settings.Debug
	slog.Debug("loading installed plugins", "path", settings.PluginsDirectory)
||||||| parent of 785acec0b (extism)
	installer.Debug = settings.Debug
	Debug("loading installed plugins from %s", settings.PluginsDirectory)
=======
	Debug("loading installed plugins from %s", settings.PluginsDirectory)
>>>>>>> 785acec0b (extism)
	plugins, err := plugin.FindPlugins(settings.PluginsDirectory)
	if err != nil {
		return err
	}
<<<<<<< HEAD
	var errorPlugins []error
||||||| parent of 785acec0b (extism)
	var errorPlugins []string
=======
	var errs []error
>>>>>>> 785acec0b (extism)

	for _, name := range o.names {
<<<<<<< HEAD
		if found := findPlugin(plugins, name); found != nil {
			if err := updatePlugin(found); err != nil {
				errorPlugins = append(errorPlugins, fmt.Errorf("failed to update plugin %s, got error (%v)", name, err))
			} else {
				fmt.Fprintf(out, "Updated plugin: %s\n", name)
			}
		} else {
			errorPlugins = append(errorPlugins, fmt.Errorf("plugin: %s not found", name))
||||||| parent of 785acec0b (extism)
		if found := findPlugin(plugins, name); found != nil {
			if err := updatePlugin(found); err != nil {
				errorPlugins = append(errorPlugins, fmt.Sprintf("Failed to update plugin %s, got error (%v)", name, err))
			} else {
				fmt.Fprintf(out, "Updated plugin: %s\n", name)
			}
		} else {
			errorPlugins = append(errorPlugins, fmt.Sprintf("Plugin: %s not found", name))
=======
		found := findPlugin(plugins, name)
		if found == nil {
			errs = append(errs, fmt.Errorf("plugin: %s not found", name))
			continue
>>>>>>> 785acec0b (extism)
		}

		if err := updatePlugin(found, out); err != nil {
			errs = append(errs, fmt.Errorf("Failed to update plugin %s: %w", name, err))
			continue
		}

		fmt.Fprintf(out, "Updated plugin: %s\n", name)
	}
<<<<<<< HEAD
	if len(errorPlugins) > 0 {
		return errors.Join(errorPlugins...)
	}
	return nil
||||||| parent of 785acec0b (extism)
	if len(errorPlugins) > 0 {
		return errors.New(strings.Join(errorPlugins, "\n"))
	}
	return nil
=======

	return errors.Join(errs...)
>>>>>>> 785acec0b (extism)
}

func updatePlugin(p *plugin.Plugin, out io.Writer) error {
	exactLocation, err := filepath.EvalSymlinks(p.Dir)
	if err != nil {
		return err
	}

	absExactLocation, err := filepath.Abs(exactLocation)
	if err != nil {
		return err
	}

	i, err := installer.FindSource(absExactLocation)
	if err != nil {
		return err
	}

	if _, err := installer.Update(i, settings); err != nil {
		return err
	}

<<<<<<< HEAD
	slog.Debug("loading plugin", "path", i.Path())
	updatedPlugin, err := plugin.LoadDir(i.Path())
	if err != nil {
		return err
	}
||||||| parent of 785acec0b (extism)
	Debug("loading plugin from %s", i.Path())
	updatedPlugin, err := plugin.LoadDir(i.Path())
	if err != nil {
		return err
	}
=======
	fmt.Fprintf(out, "Updated plugin: %s\n", p.Metadata.Name)
>>>>>>> 785acec0b (extism)

	return nil
}
