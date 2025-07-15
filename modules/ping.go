package modules

import (
	"strconv"
	"time"

	"github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/config/helpers"
)

func pingHandler(m *telegram.NewMessage) error {
	m.Delete()

	if IsSupergroup(m) {
		if !helpers.WarnIfLackOfPms(m.Client, m, m.ChatID()) {
			return telegram.EndGroup
		}
	}

	start := time.Now()
	reply, err := m.Respond("🏓 Pinging...")
	if err != nil {
		return L(m, "Modules -> ping -> Respond", err)
	}

	latency := time.Since(start).Milliseconds()
	uptime := time.Since(config.StartTime)
	uptimeStr := helpers.FormatUptime(uptime)

	text := "🏓 Pong!\n" +
		"Latency: " + strconv.Itoa(int(latency)) + "ms\n" +
		"🤖 I've been running for " + uptimeStr + " without rest!"

	reply.Edit(text)
	return telegram.EndGroup
}
