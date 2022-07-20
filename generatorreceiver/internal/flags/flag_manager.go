package flags

type Flag struct {
	Name string
	state float64
}

func (g *Flag) Enabled() bool {
	return g.state == 0
}

type FlagManager struct {
	Flags []*Flag
}

func NewFlagManager() *FlagManager {
	return &FlagManager{
		Flags: make([]*Flag, 0),
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
