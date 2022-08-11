package flags

import (
	"fmt"
	"go.uber.org/zap"
	"sync"
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
		f = f.parent()
	}
	return nil, fmt.Errorf("cyclical flag graph detected: %s", printFlagCycle(orderedFlags, f.Name()))
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
