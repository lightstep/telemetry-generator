package flags

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"log"
	"os"
	"time"
)

const (
	DisabledState = 0.0
	EnabledState  = 1.0
)

// TODO: separate config types from code types generally
// TODO: make Flag an interface

type Flag struct {
	Name          string `json:"name" yaml:"name"`
	IncidentName  string `json:"incident" yaml:"incident"`
	Start         string `json:"start" yaml:"start"`
	End           string `json:"end" yaml:"end"`
	state         float64
	cron          *cron.Cron
	incident      *Incident
	incidentStart time.Duration
	incidentEnd   time.Duration
}

func (f *Flag) Enabled() bool {
	return f.GetState() > DisabledState
}

func (f *Flag) GetState() float64 {
	if f.incident == nil {
		return f.state
	}

	// TODO: all this logic probably doesn't belong here
	if f.Name == fmt.Sprintf("%s.default", f.incident.name) {
		if !f.incident.Active() {
			return EnabledState
		}
		return DisabledState
	}
	duration := f.incident.CurrentDuration()
	if duration > f.incidentStart {
		if f.incidentEnd == 0 || duration < f.incidentEnd {
			return EnabledState
		}
	}
	return DisabledState
}

func (f *Flag) Enable() {
	f.SetState(EnabledState)
}

func (f *Flag) Disable() {
	f.SetState(DisabledState)
}

func (f *Flag) SetState(state float64) {
	f.state = state
}

func (f *Flag) Setup(im *IncidentManager, logger *zap.Logger) {
	if f.IncidentName != "" {
		f.SetupIncident(im, logger)
	} else {
		f.SetupCron(logger)
	}

}
func (f *Flag) SetupCron(logger *zap.Logger) {
	f.cron = cron.New(
		cron.WithLogger(
			cron.PrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))))
	_, err := f.cron.AddFunc(f.Start, func() {
		logger.Info("toggling flag on", zap.String("flag", f.Name))
		f.Enable()
	})
	if err != nil {
		logger.Error("error adding flag start schedule", zap.Error(err))
	}

	_, err = f.cron.AddFunc(f.End, func() {
		logger.Info("toggling flag off", zap.String("flag", f.Name))
		f.Disable()
	})
	if err != nil {
		logger.Error("error adding flag stop schedule", zap.Error(err))
	}
}

func (f *Flag) SetupIncident(im *IncidentManager, logger *zap.Logger) {
	f.incident = im.GetIncident(f.IncidentName)
	var err error
	f.incidentStart, err = time.ParseDuration(f.Start)
	if err != nil {
		logger.Error("Error setting start time")
	}
	if f.End != "" {
		f.incidentEnd, err = time.ParseDuration(f.End)
		if err != nil {
			logger.Error("Error setting start time")
		}
	}
}

func (f *Flag) StartCron() {
	if f.cron != nil {
		f.cron.Start()
	}
}

func (f *Flag) StopCron() {
	if f.cron != nil {
		f.cron.Stop()
	}
}
