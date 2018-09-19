package v1

type Workload interface {
	Containers() *WorkloadContainers
}

type WorkloadContainers struct {
	Main     Container            `json:"main"`
	Sidecars map[string]Container `json:"sidecars"`
}
