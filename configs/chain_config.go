package configs

import "math/rand"

type Chain struct {
	ChainID int    `json:"chainId"`
	Name    string `json:"name"`
	RPC     string `json:"rpc"`
	ScanURL string `json:"scanUrl"`
}

var (
	EclipseChain = Chain{
		ChainID: 9286185,
		Name:    "Eclipse",
		RPC:     "https://rpc.eclipse.io",
		ScanURL: "https://eclipsescan.xyz/tx/",
	}

	BaseChain = Chain{
		ChainID: 8453,
		Name:    "Base",
		RPC:     "https://mainnet.base.org",
		ScanURL: "https://basescan.org/tx/",
	}

	ArbitrumChain = Chain{
		ChainID: 42161,
		Name:    "Arbitrum",
		RPC:     "https://rpc.ankr.com/arbitrum",
		ScanURL: "https://arbiscan.io/tx/",
	}

	LineaChain = Chain{
		ChainID: 59144,
		Name:    "Linea",
		RPC:     "https://linea.drpc.org",
		ScanURL: "https://lineascan.build/tx/",
	}

	OptimismChain = Chain{
		ChainID: 10,
		Name:    "Optimism",
		RPC:     "https://rpc.ankr.com/optimism",
		ScanURL: "https://optimistic.etherscan.io/tx/",
	}

	ScrollChain = Chain{
		ChainID: 534352,
		Name:    "Scroll",
		RPC:     "https://rpc.ankr.com/scroll",
		ScanURL: "https://scrollscan.com/tx/",
	}

	ZkSyncChain = Chain{
		ChainID: 324,
		Name:    "ZkSync",
		RPC:     "https://rpc.ankr.com/zksync_era",
		ScanURL: "https://explorer.zksync.io/tx/",
	}
)

var chains = []Chain{
	EclipseChain,
	BaseChain,
	ArbitrumChain,
	LineaChain,
	OptimismChain,
	ScrollChain,
	ZkSyncChain,
}

func GetChainByName(chainName string) *Chain {
	for _, chain := range chains {
		if chain.Name == chainName {
			return &chain
		}
	}
	return nil
}

func GetRandomChainFromNames(chainNames []string) Chain {
	if len(chainNames) == 0 {
		return Chain{}
	}

	randomName := chainNames[rand.Intn(len(chainNames))]

	if chain := GetChainByName(randomName); chain != nil {
		return *chain
	}

	return Chain{}
}
