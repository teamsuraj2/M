package modules

import (
	"errors"

	"github.com/amarnathcjd/gogram/telegram"
)

func OnMessageFnc(m *telegram.NewMessage) error {
	if _, ok := commandSet[m.GetCommand()]; ok {
		return nil
	}

	handlers := []func(*telegram.NewMessage) error{
		DeleteAbuseHandle,
		deleteLongMessage,
		deleteLinkMessage,
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
