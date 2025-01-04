package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sync"
	"unicode/utf16"
)

type MessageWithEntity struct {
	Text     string
	Entities []tgbotapi.MessageEntity
}

type Notifier struct {
	bot            *tgbotapi.BotAPI
	chatID         int64
	walletMessages map[string][]MessageWithEntity
	mutex          sync.Mutex
}

func NewNotifier(botToken string, chatID int64) (*Notifier, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации бота: %w", err)
	}

	return &Notifier{
		bot:            bot,
		chatID:         chatID,
		walletMessages: make(map[string][]MessageWithEntity),
	}, nil
}

func (t *Notifier) AddMessageForWallet(walletAddress string, message string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if _, exists := t.walletMessages[walletAddress]; !exists {
		t.walletMessages[walletAddress] = make([]MessageWithEntity, 0)
	}
	t.walletMessages[walletAddress] = append(t.walletMessages[walletAddress], MessageWithEntity{
		Text: message,
	})
}

func (t *Notifier) AddSuccessMessageWithTxLink(walletAddress string, message string, scanUrl string, sig string) {
	fullMessage := fmt.Sprintf("✅ %s link", message)

	utf16Text := utf16.Encode([]rune(fullMessage))
	linkStart := len(utf16Text) - 4

	entity := tgbotapi.MessageEntity{
		Type:   "text_link",
		URL:    scanUrl + sig,
		Offset: linkStart,
		Length: 4,
	}

	t.mutex.Lock()
	if _, exists := t.walletMessages[walletAddress]; !exists {
		t.walletMessages[walletAddress] = make([]MessageWithEntity, 0)
	}
	t.walletMessages[walletAddress] = append(t.walletMessages[walletAddress], MessageWithEntity{
		Text:     fullMessage,
		Entities: []tgbotapi.MessageEntity{entity},
	})
	t.mutex.Unlock()
}

func (t *Notifier) SendWalletMessages(walletAddress string) error {
	t.mutex.Lock()
	messages, exists := t.walletMessages[walletAddress]
	if exists {
		delete(t.walletMessages, walletAddress)
	}
	t.mutex.Unlock()

	if !exists || len(messages) == 0 {
		return nil
	}

	var fullText string
	var allEntities []tgbotapi.MessageEntity
	currentOffset := 0

	for i, msg := range messages {
		if i > 0 {
			if i == 1 {
				fullText += "\n\n"
			} else {
				fullText += "\n"
			}
			currentOffset = len(utf16.Encode([]rune(fullText)))
		}

		fullText += msg.Text

		for _, entity := range msg.Entities {
			newEntity := entity
			newEntity.Offset = currentOffset + entity.Offset
			allEntities = append(allEntities, newEntity)
		}
	}

	msg := tgbotapi.NewMessage(t.chatID, fullText)
	msg.Entities = allEntities

	_, err := t.bot.Send(msg)
	return err
}

func (t *Notifier) ClearAllMessages() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.walletMessages = make(map[string][]MessageWithEntity)
}

func (t *Notifier) AddSuccessMessage(walletAddress string, message string) {
	t.AddMessageForWallet(walletAddress, "✅ "+message)
}

func (t *Notifier) AddErrorMessage(walletAddress string, message string) {
	t.AddMessageForWallet(walletAddress, "❌ "+message)
}
