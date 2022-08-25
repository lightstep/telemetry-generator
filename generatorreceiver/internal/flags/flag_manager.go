package flags

import (
	"fmt"
	"go.uber.org/zap"
	"sync"
	"time"
)

type FlagManager struct {
	flags map[string]*Flag

	mu sync.Mutex
}

var Manager *FlagManager

func init() {
	Manager = NewFlagManager()
}

func NewFlagManager() *FlagManager {
	return &FlagManager{flags: make(map[string]*Flag)}
}

func (fm *FlagManager) Clear() {
	fm.mu.Lock()
	fm.flags = make(map[string]*Flag)
	fm.mu.Unlock()
}

func (fm *FlagManager) GetFlags() map[string]*Flag {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	return fm.flags
}

func (fm *FlagManager) LoadFlags(configFlags []FlagConfig, logger *zap.Logger) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	for _, cfg := range configFlags {
		flag := NewFlag(cfg)
		flag.Setup(logger)
		fm.flags[flag.Name()] = &flag
	}
}

func (fm *FlagManager) ValidateFlags() error {
	validatedFlags := make(map[string]bool)
	for _, f := range fm.GetFlags() {
		if !validatedFlags[f.Name()] {
			flagGraph, err := fm.traverseFlagGraph(f)
			if err != nil {
				return err
			}
			for k, v := range flagGraph { // we know these flags are valid, so don't re-check
				validatedFlags[k] = v
			}
		}
	}
	return nil
}

func (fm *FlagManager) traverseFlagGraph(f *Flag) (map[string]bool, error) {
	seenFlags := make(map[string]bool)
	var orderedFlags []string // needed for printing flags in-order if cycle is detected

	for !seenFlags[f.Name()] {
		seenFlags[f.Name()] = true
		orderedFlags = append(orderedFlags, f.Name())
		if !f.parentSpecified() { // no parent specified -> this is a root flag, so we've traversed graph without finding cycle
			return seenFlags, nil
		}
		if f.parent() == nil { // parent was specified but it's not an actual flag
			return nil, fmt.Errorf("flag %s has parent %s which does not exist", f.Name(), f.cfg.Incident.ParentFlag)
		}
		// this is a child flag, so check that its start times and duration are valid
		err := validateChildFlagTiming(f.cfg.Incident)
		if err != nil {
			return nil, fmt.Errorf("error with flag %s: %v", f.Name(), err)
		}

		f = f.parent()
	}
	return nil, fmt.Errorf("cyclical flag graph detected: %s", printFlagCycle(orderedFlags, f.Name()))
}

func validateChildFlagTiming(incidentCfg *IncidentConfig) error {
	if len(incidentCfg.Start) == 0 {
		return fmt.Errorf("start cannot be empty")
	}
	if incidentCfg.Duration == 0 && len(incidentCfg.Start) > 1 {
		return fmt.Errorf("if duration is not specified, only one start time is permitted (will last until end of incident)")
	}
	currentStart := time.Duration(-1)
	for i := range incidentCfg.Start {
		if incidentCfg.Start[i] <= currentStart {
			return fmt.Errorf("start times must be in strictly increasing order")
		}
		currentStart = incidentCfg.Start[i]
	}
	return nil
}

func printFlagCycle(seenNames []string, repeated string) (s string) {
	for _, f := range seenNames {
		s += fmt.Sprintf("%s -> ", f)
	}
	return s + repeated
}

func (fm *FlagManager) FlagCount() int {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	return len(fm.flags)
}

func (fm *FlagManager) GetFlag(name string) *Flag {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	return fm.flags[name]
}
