package flags

import (
	"log"
	"os"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Flag struct {
	Name  string `json:"name" yaml:"name"`
	Start string `json:"start" yaml:"start"`
	End   string `json:"end" yaml:"end"`
	state float64
	cron  *cron.Cron
}

func (g *Flag) Enabled() bool {
	return g.state > 0
}

type FlagManager struct {
	Flags map[string]*Flag
}

const DISABLED_STATE = -1.0
const ENABLED_STATE = 1.0

func NewFlagManager(configFlags []Flag, logger *zap.Logger) *FlagManager {
	fm := FlagManager{make(map[string]*Flag)}

	for _, f := range configFlags {
		fm.AddFlag(f, logger)
	}

	return &fm
}

func (g *FlagManager) AddFlag(f Flag, logger *zap.Logger) {
	f.cron = cron.New(
		cron.WithLogger(
			cron.PrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))))
	_, err := f.cron.AddFunc(f.Start, func() {
		logger.Info("toggling flag on", zap.String("flag", f.Name))
		g.Enable(f.Name)
	})
	if err != nil {
		logger.Error("error adding flag start schedule", zap.Error(err))
	}

	_, err = f.cron.AddFunc(f.End, func() {
		logger.Info("toggling flag off", zap.String("flag", f.Name))
		g.Disable(f.Name)
	})
	if err != nil {
		logger.Error("error adding flag stop schedule", zap.Error(err))
	}

	g.Flags[f.Name] = &f
}

func (g *FlagManager) Start() {
	for _, v := range g.Flags {
		v.cron.Start()
	}
}

func (g *FlagManager) Stop() {
	for _, v := range g.Flags {
		v.cron.Stop()
	}
}

func (g *FlagManager) GetFlag(name string) *Flag {
	return g.Flags[name]
}

func (g *FlagManager) Enable(name string) {
	// TODO: de-hybridize this enabled / state thing
	g.Flags[name].state = ENABLED_STATE
}

func (g *FlagManager) Disable(name string) {
	// TODO: de-hybridize this enabled / state thing
	g.Flags[name].state = DISABLED_STATE

}
