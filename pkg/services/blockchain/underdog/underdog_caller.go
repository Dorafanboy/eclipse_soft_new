package underdog

import (
	"context"
	"eclipse/model"
	"eclipse/utils/requester"
	"fmt"
	"github.com/gagliardetto/solana-go/rpc"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type Module struct{}

func (m *Module) Execute(ctx context.Context, httpClient http.Client, client *rpc.Client, acc *model.EclipseAccount, words []string, maxAttempts int) (bool, error) {
	log.Println("Начал выполнение модуля Underdog Create Collection")
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

		log.Printf("Creating new collection:")
		log.Printf("- Name: %s", collection.Name)
		log.Printf("- Description: %s", collection.Description)
		log.Printf("- Image: %s", collection.Image)
		log.Printf("- Flags: Soulbound=%v, Transferable=%v, Burnable=%v",
			collection.Soulbound,
			collection.Transferable,
			collection.Burnable,
		)

		res := CreateCollection(httpClient, collection)

		err := SendSolanaTransaction(ctx, client, res, acc.PrivateKey)
		if err != nil {
			log.Printf("error creating collection from tx (попытка %d/%d): %v", attempt+1, maxAttempts, err)
			time.Sleep(3 * time.Second)
			continue
		} else {
			return true, nil
		}
	}

	return true, nil
}
