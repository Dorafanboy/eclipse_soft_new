package file

import (
	"bufio"
	"eclipse/model"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gagliardetto/solana-go"
	"os"
	"strings"
)

type AccountType int

type WordLists struct {
	Words []string
}

const (
	EVM AccountType = iota
	ECLIPSE
)

func ReadAccounts(path string, accType AccountType) (interface{}, error) {
	lines, err := ReadLines(path)
	if err != nil {
		return nil, err
	}

	switch accType {
	case EVM:
		return parseEvmAccounts(lines), nil
	case ECLIPSE:
		return parseEclipseAccounts(lines), nil
	default:
		return nil, fmt.Errorf("unknown account type")
	}
}

func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file: %w", err)
	}

	return lines, nil
}

func parseEvmAccounts(lines []string) []*model.EvmAccount {
	accs := make([]*model.EvmAccount, 0, len(lines))
	for _, line := range lines {
		if acc := createEvmAccount(line); acc != nil {
			accs = append(accs, acc)
		}
	}
	return accs
}

func parseEclipseAccounts(lines []string) []*model.EclipseAccount {
	accs := make([]*model.EclipseAccount, 0, len(lines))
	for _, line := range lines {
		if acc := createEclipseAccount(line); acc != nil {
			accs = append(accs, acc)
		}
	}
	return accs
}

func createEvmAccount(line string) *model.EvmAccount {
	if len(line) == 42 {
		address := common.HexToAddress(line)
		return model.NewEvmAccount(address, nil)
	} else if len(line) == 66 || len(line) == 64 {
		if line[0:2] == "0x" {
			line = line[2:]
		}
		privateKey, err := crypto.HexToECDSA(line)
		if err != nil {
			return nil
		}
		address := crypto.PubkeyToAddress(privateKey.PublicKey)
		return model.NewEvmAccount(address, privateKey)
	}
	return nil
}

func createEclipseAccount(line string) *model.EclipseAccount {
	privateKey, err := solana.PrivateKeyFromBase58(line)
	if err != nil {
		return nil
	}
	publicKey := privateKey.PublicKey()
	return model.NewEclipseAccount(publicKey, privateKey)
}

func LoadWordsFromFile(filepath string) (*WordLists, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			words = append(words, word)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &WordLists{Words: words}, nil
}
