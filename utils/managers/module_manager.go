package managers

import (
	"eclipse/configs"
	"eclipse/pkg/interfaces"
	"eclipse/pkg/services/blockchain/invariant"
	"eclipse/pkg/services/blockchain/lifinity"
	"eclipse/pkg/services/blockchain/orca"
	"eclipse/pkg/services/blockchain/relay"
	"eclipse/pkg/services/blockchain/solar"
	"eclipse/pkg/services/blockchain/underdog"
)

type ModuleManager struct {
	EnabledModules map[string]interfaces.ModuleInfo
	ModuleCount    int
}

func NewModuleManager(cfg configs.ModulesConfig) *ModuleManager {
	enabledModules := make(map[string]interfaces.ModuleInfo)
	moduleCount := 0

	if cfg.Enabled.Orca {
		enabledModules["Orca"] = interfaces.ModuleInfo{
			Module: &orca.Module{},
			Type:   interfaces.OrcaType,
		}
		moduleCount++
	}

	if cfg.Enabled.Underdog {
		enabledModules["Underdog"] = interfaces.ModuleInfo{
			Module: &underdog.Module{},
			Type:   interfaces.UnderdogType,
		}
		moduleCount++
	}

	if cfg.Enabled.Invariant {
		enabledModules["Invariant"] = interfaces.ModuleInfo{
			Module: &invariant.Module{},
			Type:   interfaces.DefaultType,
		}
		moduleCount++
	}

	if cfg.Enabled.Relay {
		enabledModules["Relay"] = interfaces.ModuleInfo{
			Module: &relay.Module{},
			Type:   interfaces.DefaultType,
		}
		moduleCount++
	}

	if cfg.Enabled.Lifinity {
		enabledModules["Lifinity"] = interfaces.ModuleInfo{
			Module: &lifinity.Module{},
			Type:   interfaces.DefaultType,
		}
		moduleCount++
	}

	if cfg.Enabled.Solar {
		enabledModules["Solar"] = interfaces.ModuleInfo{
			Module: &solar.Module{},
			Type:   interfaces.DefaultType,
		}
		moduleCount++
	}

	return &ModuleManager{
		EnabledModules: enabledModules,
		ModuleCount:    moduleCount,
	}
}
