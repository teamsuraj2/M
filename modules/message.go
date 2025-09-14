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

	handlers := []func(*telegram.NewMessage) error{
		deleteLongMessage,
		deleteLinkMessage,
DeleteAbuseHandle,
		deleteUserMsgIfBio,
	}

	for _, handler := range handlers {
		if err := handler(m); err != nil {
			if errors.Is(err, telegram.EndGroup) {
				return telegram.EndGroup
			}
			return L(m, "Modules -> message -> Random", err)

		}
	}

	return nil
}
