package flags

import (
	"fmt"
	"github.com/lightstep/demo-environment/generatorreceiver/internal/cron"
	"go.uber.org/zap"
	"sync"
	"time"
)

// TODO: separate config types from code types generally

type IncidentConfig struct {
	ParentFlag string          `json:"parentFlag" yaml:"parentFlag"`
	Start      []time.Duration `json:"start" yaml:"start"`
	Duration   time.Duration   `json:"duration" yaml:"duration"`
}

type CronConfig struct {
	Start string `json:"start" yaml:"start"`
	End   string `json:"end" yaml:"end"`
}

type FlagConfig struct {
	Name     string          `json:"name" yaml:"name"`
	Incident *IncidentConfig `json:"incident" yaml:"incident"`
	Cron     *CronConfig     `json:"cron" yaml:"cron"`
}

type Flag struct {
	cfg     FlagConfig
	started time.Time
	updated time.Time
	mu      sync.Mutex
}

func NewFlag(cfg FlagConfig) Flag {
	return Flag{cfg: cfg}
}

func (f *Flag) Name() string {
	return f.cfg.Name
}

func (f *Flag) Active() bool {
	f.update()
	return f.active()
}

func (f *Flag) active() bool {
	return !f.started.IsZero()
}

// update checks if the given flag f has a parent flag ("Incident"); if so,
// updates f's state based on its start and end times relative to the parent.
func (f *Flag) update() {
	if !f.parentSpecified() {
		// managed by cron or manual-only
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	parent := f.parent() // won't be nil because we already validated all parents exist
	incidentDuration := parent.CurrentDuration()
	if f.active() != f.shouldBeActive(incidentDuration) {
		f.Toggle()
	}
}

func (f *Flag) shouldBeActive(incidentDuration time.Duration) bool {
	childDuration := f.cfg.Incident.Duration
	for _, start := range f.cfg.Incident.Start {
		if incidentDuration <= start { // relies on start times being in increasing order (verified in flag validation)
			return false
		}
		// if childDuration not specified, then child flag stays active until end of incident
		if incidentDuration > start && (incidentDuration < start+childDuration || childDuration == 0) {
			return true
		}
	}
	return false
}

func (f *Flag) CurrentDuration() time.Duration {
	if !f.active() {
		return 0
	}
	return time.Since(f.started)
}

func (f *Flag) Enable() {
	if !f.active() {
		f.started = time.Now()
		f.updated = time.Now()
	}
}

func (f *Flag) Disable() {
	if f.active() {
		f.started = time.Time{}
		f.updated = time.Now()
	}
}

func (f *Flag) Toggle() {
	if f.active() {
		f.Disable()
	} else {
		f.Enable()
	}
}

func (f *Flag) Setup(logger *zap.Logger) {
	// TODO: add validation to disallow having cron and incident both configured?
	if f.cfg.Cron != nil {
		f.SetupCron(logger)
	}
}

func (f *Flag) SetupCron(logger *zap.Logger) {
	_, err := cron.Add(f.cfg.Cron.Start, func() {
		logger.Info("toggling flag on", zap.String("flag", f.cfg.Name))
		f.Enable()
	})
	if err != nil {
		logger.Error("error adding flag start schedule", zap.Error(err))
	}

	_, err = cron.Add(f.cfg.Cron.End, func() {
		logger.Info("toggling flag off", zap.String("flag", f.cfg.Name))
		f.Disable()
	})
	if err != nil {
		logger.Error("error adding flag stop schedule", zap.Error(err))
	}
}

func (f *Flag) parentSpecified() bool {
	return f.cfg.Incident != nil
}

func (f *Flag) parent() *Flag {
	if !f.parentSpecified() {
		return nil
	}
	return Manager.GetFlag(f.cfg.Incident.ParentFlag)
}

func (ic IncidentConfig) validate() error {
	if Manager.GetFlag(ic.ParentFlag) == nil {
		return fmt.Errorf("parent flag %s does not exist", ic.ParentFlag)
	}
	if len(ic.Start) == 0 {
		return fmt.Errorf("start cannot be empty")
	}
	if ic.Duration == 0 && len(ic.Start) > 1 {
		return fmt.Errorf("if duration is not specified, only one start time is permitted (will last until end of incident)")
	}
	previousStart := time.Duration(-1)
	for _, start := range ic.Start {
		if start <= previousStart {
			return fmt.Errorf("start times must be in strictly increasing order")
		}
		previousStart = start
	}
	return nil
}
