package modules

import (
	"log"
"fmt"

	"github.com/amarnathcjd/gogram/telegram"
)

func BroadcastFunc(m *telegram.NewMessage) error {
	userChan, chatChan, err := m.Client.Broadcast()
	if err != nil {
		log.Println(err)
		m.Reply(err.Error())
		return telegram.EndGroup
	}

	userCount := 0
	for range userChan {
		userCount++
	}

	chatCount := 0
	for range chatChan {
		chatCount++
	}
	m.Delete()
	m.Respond(fmt.Sprintf("Total Chats: %d\nTotal Users: %d", chatCount, chatCount))
	m.Respond("Soon implemented....")
	return telegram.EndGroup
}
