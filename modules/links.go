package modules

import (
	"fmt"

	"github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/config/helpers"
)

func deleteLinkMessage(m *telegram.NewMessage) error {
	if bo := IsSupergroup(m); !bo {
		return nil
	}

	var hasURL bool
	for _, p := range m.Message.Entities {
		if _, ok := p.(*telegram.MessageEntityURL); ok {
			hasURL = true
			break
		}
	}

	if !hasURL {
		return nil
	}

	if ShouldIgnoreGroupAnonymous(m) {
		return nil
	}
	if isadmin, err := helpers.IsChatAdmin(m.Client, m.ChannelID(), m.Sender.ID); err != nil {
		return err
	} else if isadmin {
		return nil
	}

	_, err := m.Delete()
	if err != nil {
		return err
	}
	_, err = m.Respond(fmt.Sprintf("⚠️ Direct URLs aren't allowed. Please format links properly, like <a href='%s'>this</a> for better readability.", config.SupportChannel))

	return err
}
