package modules

import (
        "fmt"
        "log"
        "regexp"

        "github.com/PaulSonOfLars/gotgbot/v2"
        "github.com/PaulSonOfLars/gotgbot/v2/ext"

        "main/config/helpers"
        "main/database"
)

func init() {
        AddHelp(
                "🛡️ BioMode",
                "biomode_help",
                "🛡️ <b>BioMode</b> monitors user bios and deletes messages if they contain URLs.\n\n"+
                        "<b>Usage:</b>\n"+
                        "➤ <code>/biolink on</code> - Enable BioMode\n"+
                        "➤ <code>/biolink off</code> - Disable BioMode\n\n"+
                        "🚫 When enabled, users with links in their bios won't be able to send messages.\n"+
                        "👮 Only admins can enable or disable this feature.",
                nil,
        )
}

func ShouldDeleteMsg(text string) bool {
        pattern := `\b(?:https?://|www\.)\S+|\b[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(?:/\S*)?|\B@\w{5,32}\b|\b\w{5,32}\.t\.me\b`
        re := regexp.MustCompile(pattern)
        return re.MatchString(text)
}

func setBioMode(m *telegram.NewMessage) error {
  if isgroup := IsValidSupergroup(m); !isgroup {
                return nil
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
        
        args := strings.Fields(m.Text())
        
        if len(args) < 2 {
                m.Respond("📚 Usage: <code>/biolink on</code> | <code>/biolink off</code>")
                return telegram.EndGroup
        }

        part := args[1]
        var msg string

        if part == "on" || part == "enable" {
                err := database.SetBioMode(m.Chat.Id)
                if err != nil {
                        msg = fmt.Sprintf("⚠️ <b>Oops! Failed to enable BioMode.</b>\n\n🚫 An error occurred while trying to turn it on.\n\n<b>Error:</b> <code>%v</code>\n\n🔁 Please try again later.", err)
                } else {
                        msg = "✅ <b>BioMode enabled successfully!</b>\n\n🔍 I will now monitor bios for any links and automatically delete messages if found.\n\n🛡 Stay safe!"
                }
        } else if part == "off" || part == "disable" {
                err := database.DelBioMode(m.Chat.Id)
                if err != nil {
                        msg = fmt.Sprintf("⚠️ <b>Oops! Failed to disable BioMode.</b>\n\n🚫 An error occurred while trying to turn it off.\n\n<b>Error:</b> <code>%v</code>\n\n🔁 Please try again later.", err)
                } else {
                        msg = "🛑 <b>BioMode disabled successfully!</b>\n\n🔓 I'm no longer monitoring user bios for links in this group.\n\n✅ You're back to normal behavior."
                }
        } else {
                msg = "❗ Invalid option. Use <code>on</code> or <code>off</code>."
        }

        m.Respond(msg)
        return telegram.EndGroup
}

func deleteUserMsgIfBio(m *telegram.NewMessage) error {
        if !IsSupergroup(m) {
		return nil
	}

        if mode, err := database.GetBioMode(m.ChatId()); err != nil {
                return err
        } else if !mode {
                return Continue
        }
if isadmin, err := helpers.IsChatAdmin(m.Client, m.ChannelID(), m.SenderID()); err != nil {
		return err
	} else if isadmin {
		return telegram.EndGroup
	}
        
      if _, ok := m.Message.FromID.(*PeerUser); !ok { return }
      
      resp , errr := m.Client.UsersGetFullUser(&telegram.InputUserObj{UserID: m.Sender.ID, AccessHash: m.Sender.AccessHash})
      if  errr != nil {
        return err
      } else resp.FullUser.About == "" || !ShouldDeleteMsg(resp.FullUser.About) {
        return nil
      }
      if _, err := m.Delete(); err != nil && handleNeedPerm(err, m) {
		return telegram.EndGroup
	}
        

                var mention string
                if m.Sender.Username != "" {
                        mention = "@" + s.User.Username
                } else {
                        mention = fmt.Sprintf("<a href='tg://user?id=%d'>%s</a>", m.Sender.Id, m.Sender.FirstName )
                }
                msg := fmt.Sprintf(`🚨 %s, your message was deleted because your bio contains a link.`, mention)
                
                return m.E(m.Respond(msg))
               
        return IsValidSupergroup(m)
}
