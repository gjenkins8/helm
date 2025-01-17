package pluginsys

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	extism "github.com/extism/go-sdk"
	"github.com/tetratelabs/wazero"
	"helm.sh/helm/v4/internal/plugins"
)

type WASMExistmPlugin struct {
	SymbolName string
	PluginWasm extism.Wasm
}

func convertExtismLogLevel(logLevel extism.LogLevel) slog.Level {
	switch logLevel {
	case extism.LogLevelDebug:
		return slog.LevelDebug
	case extism.LogLevelInfo:
		return slog.LevelInfo
	case extism.LogLevelWarn:
		return slog.LevelWarn
	case extism.LogLevelError:
		return slog.LevelError
	}
	return slog.LevelError // Default to error
}

type pluginLogger struct {
	ctx context.Context
}

func (l pluginLogger) Log(logLevel extism.LogLevel, s string) {
	// Skip output for these:
	switch logLevel {
	case extism.LogLevelOff:
		return
	case extism.LogLevelTrace: // TODO: In the future, support ability to trace if desired
		return
	}

	slog.Log(l.ctx, convertExtismLogLevel(logLevel), s)
	//fmt.Printf("%s %s: %s\n", time.Now().Format(time.RFC3339), logLevel.String(), s)
}

func createExtismPlugin(ctx context.Context, pluginManifest PluginManifest, path string) (*extism.Plugin, error) {
	//pluginBytes, err := os.ReadFile("../gotemplate-renderer.wasm")
	//require.Nil(t, err)

	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmFile{
				Path: path,
				Hash: pluginManifest.PluginSha256Hash,
				Name: pluginManifest.PluginName,
			},
			//extism.WasmData{
			//	Data: pluginBytes,
			// Name: pluginManifest.PluginName,
			//},
		},
		Memory: &extism.ManifestMemory{
			MaxPages: 65535,
			//MaxHttpResponseBytes: 1024 * 1024 * 10,
			//MaxVarBytes:          1024 * 1024 * 10,
		},
		Config: map[string]string{},
		//AllowedHosts: []string{"ghcr.io"},
		AllowedPaths: map[string]string{},
		Timeout:      0,
	}

	config := extism.PluginConfig{
		ModuleConfig:  wazero.NewModuleConfig().WithSysWalltime(),
		RuntimeConfig: wazero.NewRuntimeConfig().WithCloseOnContextDone(false),
		EnableWasi:    true,
		//EnableHttpResponseHeaders: true,
		//ObserveAdapter: ,
		//ObserveOptions: &observe.Options{},
	}
	plugin, err := extism.NewPlugin(ctx, manifest, config, []extism.HostFunction{})
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize plugin: %w", err)
	}

	plugin.SetLogger(pluginLogger{ctx}.Log)

	return plugin, nil
}

func (p *WASMExistmPlugin) Invoke(input any, output any) error {
	inputData, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	exitCode, outputData, err := plugin.Call(p.SymbolName, inputData)
	if err := p.plugin.GetError(); err != nil {
		return fmt.Errorf("failed to execute plugin: %w", err)
	}

	if exitCode != uint32(0) {
		return fmt.Errorf("plugin exited with non-zero exit code: %d", exitCode)
	}

	if err := json.Unmarshal(outputData, output); err != nil {
		return fmt.Errorf("%w", err)
	}
}
