package modules

import (
	"fmt"

	"github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/config/helpers"
)

func settingsComm(m *telegram.NewMessage) error {
	if !IsValidSupergroup(m) {
		return telegram.EndGroup
	}
	if _, err := m.Delete(); err != nil && handleNeedPerm(err, m) {
		return telegram.EndGroup
	} else if err != nil {
		return L(m, "Modules -> settings -> m.Delete()", err)
	}
	if isadmin, err := helpers.IsChatAdmin(m.Client, m.ChatID(), m.SenderID()); err != nil {
		return L(m, "Modules -> settings-> helpers.IsChatAdmin()", err)
	} else if !isadmin {
		m.Respond("Access denied: Only group admins can use this command.")

		return telegram.EndGroup
	}

	btn := telegram.NewKeyboard()
	btn.AddRow(
		telegram.Button.URL(
			"⚙️ Settings",
			fmt.Sprintf("%s?startapp=%d", config.WebAppUrl, m.ChatID()),
		),
	)
	m.Respond("Configure Your chat settings by this WebApp Interface", telegram.SendOptions{
		ReplyMarkup: btn.Build(),
	})
	return nil
}
