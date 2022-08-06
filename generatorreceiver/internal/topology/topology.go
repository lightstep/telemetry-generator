package topology

type Topology struct {
	Services []ServiceTier `json:"services" yaml:"services"`
}

func (t *Topology) GetServiceTier(serviceName string) *ServiceTier {
	for _, v := range t.Services {
		if v.ServiceName == serviceName {
			return &v
		}
	}
	return nil
}
