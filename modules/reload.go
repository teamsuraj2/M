package modules

import (
	"fmt"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/config/helpers"
)

func ReloadHandler(m *telegram.NewMessage) error {
	m.Delete()
	if !IsValidSupergroup(m) {
		return telegram.EndGroup
	}

	x, err := m.Respond("Refreshing Cache of chat admins...")
	if err != nil {
		return err
	}
	admins, _, err := m.Client.GetChatMembers(m.ChannelID(), &telegram.ParticipantOptions{
		Filter: &telegram.ChannelParticipantsAdmins{},
		Limit:  -1,
	})
	if err != nil {
		if strings.Contains(err.Error(), "CHAT_ADMIN_REQUIRED") {
			x.Edit("⚠️ I am not an admin in this group. Please promote me with necessary permissions.")
			return telegram.EndGroup
		}
		x.Edit(fmt.Sprintf("⚠️ Cache refresh failed — %v", err))
		return err
	}
	config.Cache.Store(fmt.Sprintf("admins:%d", m.ChannelID()), admins)

	var text string
	if isb, err := helpers.IsChatAdmin(m.Client, m.ChannelID(), m.SenderID()); err != nil {
		return err
	} else if isb {
		text = "✅ Successfully refreshed the cache of chat admins!"
	} else {
		text = "⚠️ Tried refreshing the admin cache... but it still seems you're not an admin!"
	}

	return m.E(x.Edit(text))
}
