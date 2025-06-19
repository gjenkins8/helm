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

package installer

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"helm.sh/helm/v4/internal/plugins/runtimes/subprocesslegacy"
	"helm.sh/helm/v4/pkg/cli"
)

type InstalledPlugin struct {
	SourceURI        string  `json:"source_uri"`
	Version          *string `json:"version"`
	PluginAPIVersion string  `json:"plugin_api_version"`
	Descriptor       string  `json:"descriptor"`
}

type InstalledPluginsIndex struct {
	APIVersion string            `json:"api_version"`
	Plugins    []InstalledPlugin `json:"plugins"`
}

// ErrMissingMetadata indicates that plugin.yaml is missing.
var ErrMissingMetadata = errors.New("plugin metadata (plugin.yaml) missing")

// Installer provides an interface for installing helm client plugins.
type Installer interface {
	// Install adds a plugin.
	Install() error
	// Path is the directory of the installed plugin.
	Path() string
	// Update updates a plugin.
	Update() error
}

// runHook will execute a plugin hook.
func runHook(p *subprocesslegacy.Plugin, event string) error {

	cmds := p.Metadata.PlatformHooks[event]
	expandArgs := true
	if len(cmds) == 0 && len(p.Metadata.Hooks) > 0 {
		cmd := p.Metadata.Hooks[event]
		if len(cmd) > 0 {
			cmds = []subprocesslegacy.PlatformCommand{{Command: "sh", Args: []string{"-c", cmd}}}
			expandArgs = false
		}
	}

	main, argv, err := subprocesslegacy.PrepareCommands(cmds, expandArgs, []string{})
	if err != nil {
		return nil
	}

	prog := exec.Command(main, argv...)

	debug("running %s hook: %s", event, prog)

	prog.Stdout, prog.Stderr = os.Stdout, os.Stderr
	if err := prog.Run(); err != nil {
		if eerr, ok := err.(*exec.ExitError); ok {
			os.Stderr.Write(eerr.Stderr)
			return errors.Errorf("plugin %s hook for %q exited with error", event, p.Metadata.Name)
		}
		return err
	}
	return nil
}

type PluginInstallHook interface {
	RunHook(event string) error
}

// Install installs a plugin.
func Install(i Installer, settings *cli.EnvSettings) (*subprocesslegacy.Plugin, error) {
	if err := os.MkdirAll(filepath.Dir(i.Path()), 0755); err != nil {
		return nil, err
	}
	if _, pathErr := os.Stat(i.Path()); !os.IsNotExist(pathErr) {
		return nil, fmt.Errorf("plugin already exists")
	}
	if err := i.Install(); err != nil {
		return nil, err
	}

	debug("loading plugin from %s", i.Path())
	p, err := subprocesslegacy.LoadDir(i.Path())
	if err != nil {
		return nil, fmt.Errorf("plugin is installed but unusable: %w", err)
	}

	subprocesslegacy.SetupPluginEnv(settings, p.Metadata.Name, p.Dir)

	if err := runHook(p, subprocesslegacy.Install); err != nil {
		return nil, err
	}

	return p, nil
}

// Update updates a plugin.
func Update(i Installer, settings *cli.EnvSettings) (*subprocesslegacy.Plugin, error) {
	if _, pathErr := os.Stat(i.Path()); errors.Is(pathErr, fs.ErrNotExist) {
		return nil, fmt.Errorf("plugin does not exist")
	}

	if err := i.Update(); err != nil {
		return nil, err
	}

	debug("loading plugin from %s", i.Path())
	p, err := subprocesslegacy.LoadDir(i.Path())
	if err != nil {
		return nil, fmt.Errorf("plugin is installed but unusable: %w", err)
	}

	subprocesslegacy.SetupPluginEnv(settings, p.Metadata.Name, p.Dir)

	if err := runHook(p, subprocesslegacy.Install); err != nil {
		return nil, err
	}

	return p, nil
}

func Uninstall(p *subprocesslegacy.Plugin) error {
	if err := os.RemoveAll(p.Dir); err != nil {
		return err
	}

	return runHook(p, subprocesslegacy.Delete)
}

// NewInstallerForSource determines the correct Installer for the given source.
func NewInstallerForSource(source, version string) (Installer, error) {
	// Check if source is a local directory
	if isLocalReference(source) {
		return NewLocalInstaller(source)
	} else if isRemoteHTTPArchive(source) {
		return NewHTTPInstaller(source)
	}
	return NewVCSInstaller(source, version)
}

// FindSource determines the correct Installer for the given source.
func FindSource(location string) (Installer, error) {
	installer, err := existingVCSRepo(location)
	if err != nil && err.Error() == "Cannot detect VCS" {
		return installer, fmt.Errorf("cannot get information about plugin source")
	}
	return installer, err
}

// isLocalReference checks if the source exists on the filesystem.
func isLocalReference(source string) bool {
	_, err := os.Stat(source)
	return err == nil
}

// isRemoteHTTPArchive checks if the source is a http/https url and is an archive
//
// It works by checking whether the source looks like a URL and, if it does, running a
// HEAD operation to see if the remote resource is a file that we understand.
func isRemoteHTTPArchive(source string) bool {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		res, err := http.Head(source)
		if err != nil {
			// If we get an error at the network layer, we can't install it. So
			// we return false.
			return false
		}

		// Next, we look for the content type or content disposition headers to see
		// if they have matching extractors.
		contentType := res.Header.Get("content-type")
		foundSuffix, ok := mediaTypeToExtension(contentType)
		if !ok {
			// Media type not recognized
			return false
		}

		for suffix := range Extractors {
			if strings.HasSuffix(foundSuffix, suffix) {
				return true
			}
		}
	}
	return false
}

// isPluginDir checks if the directory contains a plugin.yaml file.
func isPluginDir(dirname string) bool {
	_, err := os.Stat(filepath.Join(dirname, subprocesslegacy.PluginFileName))
	return err == nil
}

func debug(format string, args ...interface{}) {
	slog.Debug(fmt.Sprintf(format, args...))
}
