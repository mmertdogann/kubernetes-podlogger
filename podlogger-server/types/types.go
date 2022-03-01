package types

type PodLoggerContainer struct {
	Name  string `json:"name,omitempty"`
	Image string `json:"image,omitempty"`
}

type PodLoggerTemplate struct {
	PodName            string               `json:"podName,omitempty"`
	ContainerSize      int                  `json:"containerSize,omitempty"`
	PodLoggerContainer []PodLoggerContainer `json:"containers,omitempty"`
}

type PodLogger struct {
	PodSize           int                 `json:"podSize,omitempty"`
	PodLoggerTemplate []PodLoggerTemplate `json:"podLoggerTemplate,omitempty"`
}
