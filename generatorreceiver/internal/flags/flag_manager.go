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
	validatedFlags := make(map[*Flag]bool)
	for _, f := range fm.GetFlags() {
		if !validatedFlags[f] { // only traverse graph if we've never seen this flag before
			flagGraph, err := fm.traverseFlagGraph(f)
			if err != nil {
				return err
			}
			for k, v := range *flagGraph {
				validatedFlags[k] = v
			}
		}
	}
	return nil
}

func (fm *FlagManager) traverseFlagGraph(f *Flag) (*map[*Flag]bool, error) {
	seenFlags := make(map[*Flag]bool)
	for {
		if seenFlags[f] { // if we encounter a flag that's been seen before, then there's a cycle
			return nil, fmt.Errorf("cyclical flag graph detected: %s", printFlagCycle(&seenFlags, f))
		}
		seenFlags[f] = true
		if !f.ParentSpecified() { // no parent specified -> this is a root flag, so we've traversed graph without finding cycle
			return &seenFlags, nil
		}
		parent := f.Parent()
		if parent == nil { // parent was specified but it's not an actual flag
			return nil, fmt.Errorf("flag %s has parent %s which does not exist", f.Name(), f.cfg.Incident.ParentFlag)
		}
		f = f.Parent()
	}
}

func printFlagCycle(seenFlags *map[*Flag]bool, repeatedFlag *Flag) string {
	var s string
	for f := range *seenFlags {
		s += fmt.Sprintf("%s -> ", f.Name())
	}
	s += repeatedFlag.Name()
	return s
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
