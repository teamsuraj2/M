package modules

import (
	"time"

	"github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/config/helpers"
)

func pingHandler(m *telegram.NewMessage) error {
	m.Delete()
	if IsSupergroup(m) {
		if !helpers.WarnIfLackOfPms(m.Client, m, m.ChannelID()) {
			return telegram.EndGroup
		}
	}

	uptime := time.Since(config.StartTime)
	uptimeStr := helpers.FormatUptime(uptime)

	_, err := m.Respond("I am alive Since: " + uptimeStr)
	return err
}
