package containerapi

// GET "containers/json"

type Container struct {
	ID      string   `json:"Id"`
	Names   []string `json:"Names"`
	Image   string   `json:"Image"`
	ImageID string   `json:"ImageID"`
	State   string   `json:"State"`
	Status  string   `json:"Status"`
}

type ContainerList []Container

// GET "containers/{id}/json"

type ContainerInspect struct {
	RestartCount int `json:"RestartCount"`
}
