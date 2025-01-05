package underdog

import (
	"context"
	"eclipse/constants"
	"eclipse/internal/logger"
	"eclipse/model"
	"eclipse/pkg/services/telegram"
	"eclipse/utils/requester"
	"fmt"
	"github.com/gagliardetto/solana-go/rpc"
	"math/rand"
	"net/http"
	"time"
)

type Module struct{}

func (m *Module) Execute(
	ctx context.Context,
	httpClient http.Client,
	client *rpc.Client,
	acc *model.EclipseAccount,
	notifier *telegram.Notifier,
	words []string,
	maxAttempts int,
) (bool, error) {
	logger.Info("Начал выполнение модуля Underdog Create Collection")
	rand.Seed(time.Now().UnixNano())

	for attempt := 0; attempt < maxAttempts; attempt++ {
		word1 := words[rand.Intn(len(words))]
		word2 := words[rand.Intn(len(words))]
		name := fmt.Sprintf("%s %s", word1, word2)

		descWord := words[rand.Intn(len(words))]
		imageUrl := requester.GetOneRandomImage(httpClient)

		collection := CollectionData{
			Account:      acc.PublicKey.String(),
			Name:         name,
			Image:        imageUrl,
			Description:  descWord,
			ExternalUrl:  "",
			Soulbound:    rand.Float32() < 0.5,
			Transferable: rand.Float32() < 0.5,
			Burnable:     rand.Float32() < 0.5,
		}

		logger.Info("Creating new collection:")
		logger.Info("- Name: %s", collection.Name)
		logger.Info("- Description: %s", collection.Description)
		logger.Info("- Image: %s", collection.Image)
		logger.Info("- Flags: Soulbound=%v, Transferable=%v, Burnable=%v",
			collection.Soulbound,
			collection.Transferable,
			collection.Burnable,
		)

		res := CreateCollection(httpClient, collection)

		sig, err := SendSolanaTransaction(ctx, client, res, acc.PrivateKey)
		if err != nil {
			logger.Error("error creating collection from tx (попытка %d/%d): %v", attempt+1, maxAttempts, err)
			time.Sleep(3 * time.Second)
			continue
		} else {
			notifier.AddSuccessMessageWithTxLink(
				acc.PublicKey.String(),
				"Underdog",
				constants.EclipseScan,
				sig.String(),
			)
			return true, nil
		}
	}

	notifier.AddErrorMessage(acc.PublicKey.String(), "Underdog")
	return false, fmt.Errorf("could not execute create collection after %d attempts", maxAttempts)
}
