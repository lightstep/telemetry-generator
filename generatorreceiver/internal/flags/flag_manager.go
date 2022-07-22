package flags

import (
	"go.uber.org/zap"
)

type FlagManager struct {
	Flags map[string]*Flag
}

func NewFlagManager(im *IncidentManager, configFlags []Flag, logger *zap.Logger) *FlagManager {
	fm := FlagManager{make(map[string]*Flag)}

	for _, f := range configFlags {
		flag := f
		flag.Setup(im, logger)
		fm.Flags[flag.Name] = &flag
	}

	return &fm
}

func (g *FlagManager) Start() {
	for _, v := range g.Flags {
		v.StartCron()
	}
}

func (g *FlagManager) Stop() {
	for _, v := range g.Flags {
		v.StopCron()
	}
}

func (g *FlagManager) GetFlag(name string) *Flag {
	return g.Flags[name]
}
