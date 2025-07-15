package modules

import (
	"fmt"

	"github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/database"
)

func botAddded(m *telegram.ParticipantUpdate) error {
	if m.UserID() != m.Client.Me().ID || (!m.IsAdded() && !m.IsKicked() && !m.IsBanned()) {
		return telegram.EndGroup
	}
	text := fmt.Sprintf(
		`Hello 👋 I'm <b>%s</b>, here to help keep the chat transparent and secure.

🚫 I will automatically delete edited messages to maintain clarity.  

I'm ready to protect this group! ✅  
Let me know if you need any help.`,
		m.Client.Me().FirstName,
	)

	if .IsAdded() {
		m.Client.SendMessage(
			m.ChannelID(),
			text,
		)
		database.AddServedChat(m.ChannelID())
	}
	if r := database.IsLoggerEnabled(); !r {
		return telegram.EndGroup
	}
	var chatMemberCount int32
	var status, groupUsername, groupTitle, logStr string
	if _, cmc, err := m.Client.GetChatMembers(m.ChannelID(), &telegram.ParticipantOptions{
		Limit: -1,
	}); err != nil {
		chatMemberCount = 0
	} else {
		chatMemberCount = cmc
	}

	if chat, err := m.Client.GetChannel(m.ChannelID()); err != nil {
		groupUsername = "N/A"
		groupTitle = "N/A"
	} else {
		if chat.Username == "" {
			groupUsername = "N/A"
		} else {
			groupUsername = "@" + chat.Username
		}
		if chat.Title == "" {
			groupTitle = "N/A"
		} else {
			groupTitle = chat.Title
		}
	}

	if m.IsAdded() {
		status = "Added"
	} else if m.IsKicked() || !m.IsBanned() {
		status = "Removed"
	}
	logStr = fmt.Sprintf(
		`🔹 <b>Bot was %s in Group</b> 🔹  
━━━━━━━━━━━━━━━━━  
📌 <b>Group Name:</b> %s  
🆔 <b>Group ID:</b> <code>%d</code>  
🔗 <b>Username:</b> %s  
👥 <b>Members:</b> %d  
━━━━━━━━━━━━━━━━━`,
		status,
		groupTitle,
		m.ChannelID(),
		groupUsername,
		chatMemberCount,
	)
	_, err := m.Client.SendMessage(
		config.LoggerId,
		logStr,
	)
	return err
}
