package modules

import (
	"errors"
	"log"
)

func OnMessageFnc(m *telegram.NewMessage) error {
	handlers := []func(*telegram.NewMessage) error{
		deleteLongMessage,
		deleteLinkMessage,
		deleteUserMsgIfBio,
	}

	for _, handler := range handlers {
		if err := handler(m); err != nil {
			if errors.Is(err, telegram.EndGroup) {
				return telegram.EndGroup
			}

			_, e := m.Client.SendMessage(telegram.LoggerId, "Error in OnMessageFnc: "+err.Error())
			if e != nil {
				log.Println("Error in OnMessagFnc error send to grp", err.Error())
			}
			return telegram.EndGroup
		}
	}

	return nil
}
