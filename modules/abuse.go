package modules

func init() {
	AddHelp(
		"🚫 NoAbuse",
		"noabuse_help",
		"<b>🚫 NoAbuse Filter</b>\n" +
			"Automatically detects and filters abusive or offensive language in group messages.\n\n" +
			"<b>🔧 Commands:</b>\n" +
			"• <code>/noabuse on</code> – Enable abuse detection ✅\n" +
			"• <code>/noabuse off</code> – Disable abuse detection ❌\n\n" +
			"<b>ℹ️ Notes:</b>\n" +
			"– Messages with offensive content will be censored or removed.\n" +
			"– 👮 Only group admins can configure this setting.",
	)
}
func AddAbuseCmd(m *telegram.NewMessage) error {
	`args := strings.Fields(m.Text())
	if !IsValidSupergroup(m) {
		return telegram.EndGroup
	}
	if _, err := m.Delete(); err != nil && handleNeedPerm(err, m) {
		return telegram.EndGroup
	}
	isAdmin, err := helpers.IsChatAdmin(m.Client, m.ChannelID(), m.SenderID())
	if err != nil {
		return err
	} else if !isAdmin {
		m.Respond("🚫 Only group admins can add abusive words.")
		return telegram.EndGroup
	}
	if len(args) < 2 {
		m.Respond("⚠️ Usage: <code>/addabuse word</code>")
		return telegram.EndGroup
	}
	if len(word) > 25 {
	  m.Respond("❌ Word too long. Keep it under 25 characters.\nUse *, **, ? for matching. See /help for details.")
	return telegram.EndGroup
	  
	}
	
	word := string.TrimSpace(strings.ToLower(args[1]))
	if word == "word" {
	  m.Respond("'m' is not a valid word, Please provide valid one.")
	}
	patterns, err := database.GetNSFWWords()
	if err != nil {
		m.Respond("❌ Failed to fetch abuse list.")
		return telegram.EndGroup
	}
	if MatchAnyPattern(patterns, word) {
		m.Respond("ℹ️ This word is already covered by the any abuse filters.")
		return telegram.EndGroup
	}
	if err := database.AddNSFWWord(word); err != nil {
		m.Respond("❌ Failed to add word.")
		return telegram.EndGroup
	}
	m.Respond("✅ Word added to abuse list: <code>" + word + "</code>")
`
	return telegram.EndGroup
}

func NoAbuseCmd(m *telegram.NewMessage) error {
	args := strings.Fields(m.Text())
	if isgroup := IsValidSupergroup(m); !isgroup {
		return telegram.EndGroup
	}
	if _, err := m.Delete(); err != nil && handleNeedPerm(err, m) {
		return telegram.EndGroup
	}
	if isadmin, err := helpers.IsChatAdmin(m.Client, m.ChannelID(), m.SenderID()); err != nil {
		return err
	} else if !isadmin {
		m.Respond("Access denied: Only group admins can use this command.")

		return telegram.EndGroup
	}

	if len(args) < 2 {
		m.Respond("⚠️ Usage:\n<code>/noabuse on</code> – Enable Abuse detection\n<code>/noabuse off</code> – Disable abuse detection")
		return telegram.EndGroup
	}
	arg := strings.ToLower(args[1])
	enable := arg == "on"
	if arg != "on" && arg != "off" {
		m.Respond("❌ Invalid option.\nUse <code>/noabuse on</code> or <code>/noabuse off</code>")
		return telegram.EndGroup
	}
	if err := database.SetNSFWFlag(m.Chat.ID, enable); err != nil {
		log.Println("NoAbuse.error:", err)
		m.Respond("❌ Failed to update setting.")
		return telegram.EndGroup
	}
	status := "🛡️ NoAbuse detection is enabled ✅"
if !enable {
	status = "🚫 NoAbuse detection is disabled ❌"
}
	m.Respond(status)
	return telegram.EndGroup
}
