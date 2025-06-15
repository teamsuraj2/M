package modules

import (
	"fmt"
	"log"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/config/helpers"
)

func handleNeedPerm(e error, m *telegram.NewMessage) bool {
 if e == nil {
return false 
} else 	if strings.Contains(e.Error(), "MESSAGE_DELETE_FORBIDDEN") {
		m.Respond("I need 'Delete Message' Permission to work properly")
		return true
	}

	log.Println("Error hanlded by HandleNeedPerm: %v", e)
	return true
}

func IsSupergroup(m *telegram.NewMessage) bool {
	return m.ChatType() == telegram.EntityChat && m.Channel != nil
}

func IsBasicGroup(m *telegram.NewMessage) bool {
	return m.ChatType() == telegram.EntityChat && m.Channel == nil
}

func IsAnonymousAdmin(m *telegram.NewMessage) bool {
	if m == nil || m.Message == nil {
		return false
	}
	if m.Message.FromID == nil {
		return true
	}
	if _, isChannel := m.Message.FromID.(*telegram.PeerChannel); isChannel {
		return true
	}
	return false
}

func IsValidSupergroup(m *telegram.NewMessage) bool {
	if !IsSupergroup(m) {
		m.Reply("⚠️ This command is only usable in supergroups!")
		return false
	}
	if IsAnonymousAdmin(m) {
		m.Reply("🚫 You are an anonymous admin. You can't use this command.")
		return false
	}
	return true
}

func ShouldIgnoreGroupAnonymous(m *telegram.NewMessage) bool {
	if !IsAnonymousAdmin(m) {
		return false
	}

	if m.Sender.ID == m.Chat.ID { // Sender is group admin
		return true
	}
	if m.Sender.ID == 0 && m.SenderID() == 0 { // Group anonmous
		return true
	}

	fullChat, err := helpers.GetFullChannel(m.Client, m.ChannelID())
	if err != nil {
		m.Client.SendMessage(config.LoggerId, fmt.Sprintf("Failed to get fullchannel for %d\n Errer: %v", m.ChannelID(), err))
		log.Println(err)
		return false
	}
	// Check if it's a linked channel message
	if m.Sender.ID == fullChat.LinkedChatID {
		return true
	}

	return false
}
