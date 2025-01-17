package pluginsys

import (
	"context"
	"fmt"
	"os"

	extism "github.com/extism/go-sdk"
)

type PluginManager struct {
}

func NewPluginManager() (*PluginManager, error) {
	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmUrl{
				Url: "https://github.com/extism/plugins/releases/latest/download/count_vowels.wasm",
			},
		},
	}

	ctx := context.Background()
	config := extism.PluginConfig{}
	p, err := extism.NewPlugin(ctx, manifest, config, []extism.HostFunction{})

	if err != nil {
		fmt.Printf("Failed to initialize plugin: %v\n", err)
		os.Exit(1)
	}

	data := []byte("Hello, World!")
	exit, out, err := p.Call("count_vowels", data)
	if err != nil {
		fmt.Println(err)
		os.Exit(int(exit))
	}

	response := string(out)
	fmt.Println(response)

	return nil, nil
}
