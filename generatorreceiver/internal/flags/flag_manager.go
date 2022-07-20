package flags

import (
	"log"
	"os"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Flag struct {
	Name string `json:"name" yaml:"name"`
	Start string `json:"start" yaml:"start"`
	End string `json:"end" yaml:"end"`
	state float64
	cron *cron.Cron
}

func (g *Flag) Enabled() bool {
	return g.state == 0
}

type FlagManager struct {
	Flags []*Flag
}

func NewFlagManager(configFlags []Flag, logger *zap.Logger) *FlagManager {
	var flags []*Flag

	for _, v := range configFlags {
		v.cron = cron.New(
			cron.WithLogger(
				cron.PrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))))
		_, err := v.cron.AddFunc(v.Start, func() {
			logger.Info("toggling flag on", zap.String("flag", v.Name))
			v.state = 1.0
		})
		if err != nil {
			logger.Error("error adding flag start schedule", zap.Error(err))
		}

		_, err = v.cron.AddFunc(v.End, func() {
			logger.Info("toggling flag off", zap.String("flag", v.Name))
			v.state = 0.0
		})
		if err != nil {
			logger.Error("error adding flag stop schedule", zap.Error(err))
		}

		flags = append(flags, &v)
	}

	return &FlagManager{
		Flags: flags,
	}
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
	for _, f := range g.Flags {
		if f.Name == name {
			return f
		}
	}
	return nil
}
