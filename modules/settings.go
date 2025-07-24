package modules

import (
	"github.com/amarnathcjd/gogram/telegram"

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

	/*webButton := telegram.Button.WebView("🌐 Open WebApp", "https://gotgboy-571497f84322.herokuapp.com/")
	markup := telegram.Button.Keyboard(
		telegram.Button.Row(webButton),
	)
	*/
	m.Respond("in progress...")
	return nil
}
