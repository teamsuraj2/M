package modules

import (
	"errors"
	"log"

	"github.com/amarnathcjd/gogram/telegram"

	"main/config"
)

func OnMessageFnc(m *telegram.NewMessage) error {
	if m.GetCommand() != "" {
		return nil
	}
	handlers := []func(*telegram.NewMessage) error {
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

			_, e := m.Client.SendMessage(config.LoggerId, "Error in OnMessageFnc: "+err.Error())
			if e != nil {
				log.Println("Error in OnMessagFnc error send to grp", err.Error())
			}
			return telegram.EndGroup
		}
	}

	return nil
}
