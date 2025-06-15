package modules

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"

	"main/config/helpers"
	"main/database"
)

func SetLongLimitHandler(m *telegram.NewMessage) error {
	m.Delete()
	if !IsValidSupergroup(m) {
		return telegram.EndGroup
	}

	if m.Sender == nil {
		return Continue
	}
	if isadmin, err := helpers.IsChatAdmin(m.Client, m.ChannelID(), m.SenderID()); err != nil {
		return err
	} else if !isadmin {
		m.Respond("Access denied: Only group admins can use this command.")

		return telegram.EndGroup
	}

	if m.Args() == "" {
		m.Respond("Usage: /setlonglimit <number>")
		return telegram.EndGroup
	}

	args := strings.Fields(m.Args())
	limit, convErr := strconv.Atoi(args[0])
	if convErr != nil || limit < 200 || limit > 4000 {
		m.Respond("Please provide a valid number between 200 and 4000.")
		return telegram.EndGroup
	}

	settings := &database.EchoSettings{
		ChatID: m.ChannelID(),
		Limit:  limit,
	}

	err := database.SetEchoSettings(settings)
	if err != nil {
		m.Respond(fmt.Sprintf("Error saving limit: %v", err))
		return telegram.EndGroup
	}

	m.Respond(fmt.Sprintf("✅ Long message limit set to %d characters.", limit))
	return telegram.EndGroup
}

func SetLongModeHandler(m *telegram.NewMessage) error {
	if isgroup := IsValidSupergroup(m); !isgroup {
		return telegram.EndGroup
	}
	if m.Sender == nil {
		return Continue
	}
	if isadmin, err := helpers.IsChatAdmin(m.Client, m.ChannelID(), m.SenderID()); err != nil {
		return err
	} else if !isadmin {
		m.Respond("Access denied: Only group admins can use this command.")

		return telegram.EndGroup
	}

	if m.Args() == "" {
		m.Respond("Usage: /setlongmode <off|manual|automatic>")
		return telegram.EndGroup
	}
	mode := strings.ToLower(strings.Fields(m.Args())[0])
	if mode != "off" && mode != "manual" && mode != "automatic" && mode != "auto" {
		m.Respond("Invalid mode. Use one of: <code>off</code>, <code>manual</code>, <code>automatic</code>.")
		return telegram.EndGroup
	}

 if mode == "auto"{
mode="automatic"
}
	settings := &database.EchoSettings{
		ChatID: m.ChannelID(),
		Mode:   mode,
	}

	err := database.SetEchoSettings(settings)
	if err != nil {
		m.Respond(fmt.Sprintf("Error saving mode: %v", err))
		return telegram.EndGroup
	}

	m.Respond(fmt.Sprintf("✅ Long message mode set to '%s'.", mode))
	return telegram.EndGroup
}
