package telebot

import (
	"time"
)

// Bot represents a separate Telegram bot instance.
type Bot struct {
	Token string

	// Bot as `User` on API level.
	Identity User
}

// Listen periodically looks for updates and delivers new messages
// to subscription channel.
func (b *Bot) Listen(subscription chan<- Message, interval time.Duration) {
	updates := make(chan Update)
	pulse := time.NewTicker(interval)
	latest_update := 0

	go func() {
		for range pulse.C {
			go api_getUpdates(b.Token,
				latest_update+1,
				updates)
		}
	}()

	go func() {
		for update := range updates {
			if update.Id > latest_update {
				latest_update = update.Id
			}

			subscription <- update.Payload
		}
	}()
}

// SendMessage sends a text message to recipient.
func (b *Bot) SendMessage(recipient User, message string) error {
	return api_sendMessage(b.Token, recipient, message)
}

// ForwardMessage forwards a message to recipient.
func (b *Bot) ForwardMessage(recipient User, message Message) error {
	return api_forwardMessage(b.Token, recipient, message)
}

// SendPhoto sends a photo object to recipient.
func (b *Bot) SendPhoto(recipient User, photo *Photo) error {
	return api_sendPhoto(b.Token, recipient, photo)
}
