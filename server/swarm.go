package server

type Node struct {
	ID             string   `json:"id,omitempty"`
	Name           string   `json:"name,omitempty"`
	Addr           string   `json:"addr,omitempty"`
	Containers     string   `json:"containers,omitempty"`
	ReservedCPUs   string   `json:"reserved_cpus,omitempty"`
	ReservedMemory string   `json:"reserved_memory,omitempty"`
	Labels         []string `json:"labels,omitempty"`
}

func (s *Server) getSwarmNodes() ([]*Node, error) {
	client, err := s.getDockerClient()
	if err != nil {
		return nil, err
	}

	info, err := client.Info()
	if err != nil {
		return nil, err
	}

	nodes, err := parseSwarmNodes(info.DriverStatus)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}
