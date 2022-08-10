package topology

type Topology struct {
	Services map[string]*ServiceTier `json:"services" yaml:"services"`
}

func (t *Topology) GetServiceTier(serviceName string) *ServiceTier {
	return t.Services[serviceName]
}

func (t *Topology) LoadTopology() error {
	for name, service := range t.Services {
		service.ServiceName = name
		err := service.loadRoutes()
		if err != nil {
			return err
		}
	}
	return nil
}
