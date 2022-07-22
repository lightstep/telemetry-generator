package flags

import "time"

type IncidentManager struct {
	incidents map[string]*Incident
}

func NewIncidentManager() *IncidentManager {
	return &IncidentManager{incidents: make(map[string]*Incident)}
}

func (im *IncidentManager) GetIncidents() []*Incident {
	out := []*Incident{}
	for _, i := range im.incidents {
		out = append(out, i)
	}
	return out
}

func (im *IncidentManager) GetIncident(name string) *Incident {
	incident, exists := im.incidents[name]
	if !exists {
		incident = &Incident{name: name}
		im.incidents[name] = incident
	}
	return incident
}

// TODO: we may need a way to get from incident to flags

type Incident struct {
	name  string
	start time.Time
}

func (i *Incident) GetName() string {
	return i.name
}

func (i *Incident) Active() bool {
	return i.CurrentDuration() > 0
}

func (i *Incident) CurrentDuration() time.Duration {
	if i.start.IsZero() {
		return time.Duration(0)
	}
	return time.Now().Sub(i.start)
}

func (i *Incident) Start() {
	if i.start.IsZero() {
		i.start = time.Now()
	}
}

func (i *Incident) End() {
	i.start = time.Time{}
}
