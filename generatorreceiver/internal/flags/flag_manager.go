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

func (fm *FlagManager) LoadFlags(configFlags []FlagConfig, logger *zap.Logger) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	for _, cfg := range configFlags {
		flag := NewFlag(cfg)
		flag.Setup(logger)
		fm.flags[flag.Name()] = &flag
	}
	err := fm.validateFlags()
	if err != nil {
		return err
	}
	return nil
}

func (fm *FlagManager) validateFlags() error {
	for name, flag := range fm.flags {
		if flag.cfg.Incident != nil {
			parentName := flag.cfg.Incident.ParentFlag
			if parentName == "" {
				return fmt.Errorf("flag %+v is associated with incident but missing parent flag", name)
			}
			parentFlag := fm.flags[parentName]
			if parentFlag == nil {
				return fmt.Errorf("flag %+v has parentFlag %v which does not exist", name, parentName)
			}
			// todo- check for loops in flag graph
		}
	}
	return nil
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
