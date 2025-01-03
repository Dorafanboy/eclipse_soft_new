package lifinity

import (
	"eclipse/internal/base"
	"github.com/gagliardetto/solana-go"
)

var (
	ETH_USDC_POOL = base.SwapConfig{
		ProgramID:     LIFINITY_PROGRAM_ID,
		PoolAddress:   solana.MustPublicKeyFromBase58("6BYe6v7TFZNNdEdAbCU47A5hQzR9eYb3ZfG4QZvM4sWg"),
		StateAddress:  solana.MustPublicKeyFromBase58("64EYB9hAmVJ9kpBNLkmaHhiVnGdr5PJEwi4K7PdxBiLH"),
		VaultA:        solana.MustPublicKeyFromBase58("4uPn7ksHrDB3hbMkReGQXzdMQmRxfWxZzLzLFnPH6j6e"),
		VaultB:        solana.MustPublicKeyFromBase58("cp5Kejxkt7rn5j6SnTFziK8VfXgjAHGra86ApSs1MrP"),
		PoolAuthority: solana.MustPublicKeyFromBase58("BJYusw5QvkWGLTwsXuVwvkjZJcMDW8SKGengPcdcgJ1m"),
		FeeAccount:    solana.MustPublicKeyFromBase58("FdZkrEFo5aE4gnGvJCwEQe6D3aTU8c2a86VncGJ77yfU"),
		FeeState:      solana.MustPublicKeyFromBase58("8LztqYZnj4YpwYf34bPA8TkmLSacqLGbdNyF94CWxXUp"),
		OracleAddress: solana.MustPublicKeyFromBase58("CYESsyLqZb5qLxmBiRaMfJzWho9uaJtHZ99kGCg7Wf8K"),
	}

	SOL_USDC_POOL = base.SwapConfig{
		ProgramID:     LIFINITY_PROGRAM_ID,
		PoolAddress:   solana.MustPublicKeyFromBase58("7M5tCMMbxW8bRTymTjJNbk1ReYrgsfzxogLwsk6vf55X"),
		StateAddress:  solana.MustPublicKeyFromBase58("8cCS9KGb1nArpEJPExfQzAk3VKWS2Zc5KdTmqYYFTkni"),
		VaultA:        solana.MustPublicKeyFromBase58("795o43yqr6qPBnbaeUKAo5rPSiouB3h9gUFbSEyn7gNY"),
		VaultB:        solana.MustPublicKeyFromBase58("AQQjZw7bohTLvBiNUF9ZBp4f41rifyfv5RcKbsuHnnHL"),
		PoolAuthority: solana.MustPublicKeyFromBase58("7YAFeKLJF1BQHwdXXMzPgjrHoHZewDK58QwcC7o5YYa"),
		FeeAccount:    solana.MustPublicKeyFromBase58("9DWRkciofDeNyj8FRjQEYeABxwsh4VQbBx5FiVvY9Cu5"),
		FeeState:      solana.MustPublicKeyFromBase58("At4Gdd43Ri4EmwteV684hAEAmsnB1ZMbvk5nX9xw8XW"),
		OracleAddress: solana.MustPublicKeyFromBase58("6Si5jzZCnZYzqU9ap8NGMjdar5z3stZJGoi7PRh7Z4hc"),
	}
)
