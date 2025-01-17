package plugintypes

import chart "helm.sh/helm/v4/pkg/chart/v2"

type RendererPluginRenderChartRequestV0 struct {
	Chart            *chart.Chart `json:"chart"`
	RenderValuesJSON []byte       `json:"values"`
}

type RendererPluginOutputManifestV0 struct {
	Filename string `json:"filename"`
	Content  []byte `json:"manifest"`
}

type RendererPluginRenderChartResponseV0 struct {
	Manifests []RendererPluginOutputManifestV0 `json:"manifests"`
}
