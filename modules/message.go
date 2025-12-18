package modules

import (
	"errors"

	"github.com/amarnathcjd/gogram/telegram"

	"main/database"
)

func OnMessageFnc(m *telegram.NewMessage) error {
	go func() {
		if m.IsPrivate() {
			database.AddServedUser(m.ChatID())
		} else {
			database.AddServedChat(m.ChatID())
		}
	}()

	if _, ok := commandSet[m.GetCommand()]; ok {
		return nil
	}

	// All message handlers
	handlers := []func(*telegram.NewMessage) error{
		// Existing handlers
		deleteLongMessage,
		deleteLinkMessage,
		DeleteAbuseHandle,
		deleteUserMsgIfBio,

		// NEW HANDLERS
		handleMediaDelete,      // Media auto-delete
		handleMsgAutoDelete,    // Message auto-delete
		handleForwardedMessage, // Forward blocking
		handlePhoneNumber,      // Phone number blocking
	}

	for _, handler := range handlers {
		if err := handler(m); err != nil {
			if errors.Is(err, telegram.EndGroup) {
				return telegram.EndGroup
			}
			return L(m, "Modules -> message -> Handler", err)
		}
	}

	return nil
}
