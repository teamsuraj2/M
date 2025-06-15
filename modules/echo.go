package modules

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/config/helpers"
	"main/database"
)

func init() {
	AddHelp("📝 Long Message", "echo", `<b>Command:</b>
<blockquote>/echo &lt;text&gt;
 /setlongmode &lt;off|manual|automatic&gt;
 /setlonglimit &lt;number&gt;</blockquote>

<b>Description:</b>
Sends back the provided text. Also allows setting how the bot handles long messages.

<b>Echo Text:</b>
• <b>/echo</b> &lt;text&gt; – If the message is too long, uploads it to Telegraph and sends as the link.
• <b>/echo</b> &lt;text&gt; (with reply) – Replies with Telegraph link instead of sending normally.

<b>Mode Settings:</b>
• <b>/setlongmode</b> <code>off</code> – No action on long messages.
• <b>/setlongmode</b> <code>manual</code> – Deletes, warns user.
• <b>/setlongmode</b> <code>automatic</code> – Deletes, sends Telegraph link ( Default ).

<b>Custom Limit:</b>
• <b>/setlonglimit</b> <code>&lt;number&gt;</code> – Set character limit (200–4000, default: 800).`)
}

type warningTracker struct {
	globalLock sync.Mutex            // Protects access to locks map
	locks      map[int64]*sync.Mutex // Per-chat locks
	chats      map[int64]time.Time   // Last warning per chat
}

var deleteWarningTracker = warningTracker{
	locks: make(map[int64]*sync.Mutex),
	chats: make(map[int64]time.Time),
}

func EcoHandler(m *telegram.NewMessage) error {
	if isgroup := IsValidSupergroup(m); !isgroup {
		return telegram.EndGroup
	}
	_, err := m.Delete()
 if x:= handleNeedPerm(err, m); x{
		return telegram.EndGroup
	}

	if m.Args() == "" {
		m.Reply("Usage: /echo &lt;long message&gt;")
		return telegram.EndGroup
	}

	m.Delete()

	settings, err = database.GetEchoSettings(m.ChannelID())
	if err != nil {
		m.Respond(fmt.Sprintf("⚠️ Something went wrong while processing the limit.\nError: %v", err))
		return err
	}

	if len(m.Text()) < settings.Limit {
		if isadmin, err := helpers.IsChatAdmin(m.Client, m.ChannelID(), m.SenderID()); err != nil {
			return err
		} else if isadmin {
			m.Respond(fmt.Sprintf("Oops! Your message is under %d characters. Since you're an admin, you're not required to use /echo. But if you’d like to, please send a message longer than %d characters.", settings.Limit, settings.Limit))
			return telegram.EndGroup
		}
		m.Respond(fmt.Sprintf("Oops! Your message is under %d characters. You can send it without using /echo.", settings.Limit))
		return telegram.EndGroup
	}

	text := strings.SplitN(m.Text(), " ", 2)[1]

	err = sendEchoMessage(m, text)
	return orCont(err)
}

func deleteLongMessage(m *telegram.NewMessage) error {
	if !IsSupergroup(m) {
		return nil
	}

	if m.GetCommand() == "/echo" {
		return nil
	}

	if ShouldIgnoreGroupAnonymous(m) {
		return nil
	}
	chatID := m.ChannelID()
	if isadmin, err := helpers.IsChatAdmin(m.Client, chatID, m.Sender.ID); err != nil {
		return err
	} else if isadmin {
		return nil
	}
	settings, err := database.GetEchoSettings(chatID)
	var isAutomatic bool
	if err != nil {
		m.Client.SendMessage(
			config.LoggerId,
			fmt.Sprintf("⚠️ Something went wrong while Getting the limit.\nError: %v", err),
		)
		return err
	}

	if m.Text() == "" || len(m.Text()) < settings.Limit {
		return nil
	}
	if settings.Mode == "OFF" {
		return nil
	}
	if settings.Mode == "AUTOMATIC" {
		isAutomatic = true
	}
	_, err = m.Delete()
	if err != nil {
		fmt.Println(" Long mode automatic Delete error:", err)
		return err
	}

	if !isAutomatic {
		deleteWarningTracker.Lock(chatID)
		defer deleteWarningTracker.Unlock(chatID)

		lastWarning, exists := deleteWarningTracker.chats[chatID]
		if !exists || time.Since(lastWarning) > time.Second {
			var name string
			var id int64
			if m.SenderChat != nil {
				name = m.SenderChat.Title
				id = m.SenderChat.ID
			} else if m.Sender != nil {
				name = strings.TrimSpace(m.Sender.FirstName + " " + m.Sender.LastName)
				id = m.Sender.ID
			}
			text := fmt.Sprintf(`
⚠️ <a href="tg://user?id=%d">%s</a>, your message exceeds the %d-character limit! 🚫  
Please shorten it before sending. ✂️  

Alternatively, use /echo for sending longer messages. 📜
`, id, name, settings.Limit)

			_, err := m.Respond(text)
			if err != nil {
				fmt.Println("echo manul SendMessage error:", err)
				return err
			}
			deleteWarningTracker.chats[chatID] = time.Now()
		}
	} else if isAutomatic {
		err = sendEchoMessage(m, m.Text())
		return orCont(err)
	}
	return nil
}

func sendEchoMessage(m *telegram.NewMessage, text string) error {
	var userFullName, authorURL string
	if m.SenderChat != nil {
		userFullName = m.SenderChat.Title
		if u := m.SenderChat.Username; u != "" {
			authorURL = fmt.Sprintf("https://t.me/%s", u)
		} else {
			authorURL = fmt.Sprintf("https://t.me/%s?start=info_%d", m.Client.Me().Username, m.SenderChat.ID)
		}
	} else if m.Sender != nil {
		userFullName = strings.TrimSpace(m.Sender.FirstName + " " + m.Sender.LastName)
		if um := m.Sender.Username; um != "" {
			authorURL = fmt.Sprintf("https://t.me/%s", um)
		} else {
			authorURL = fmt.Sprintf("https://t.me/%s?start=info_%d", m.Client.Me().Username, m.Sender.ID)
		}
	}

	url, err := helpers.CreateTelegraphPage(text, userFullName, authorURL)
	if err != nil {
		return err
	}

	msgTemplate := `Hello <a href="tg://user?id=%d">%s</a>, <a href="tg://user?id=%d">%s</a> wanted to share a message, but it was too long to send here. You can view the full message on <b><a href="%s">Telegraph 📝</a></b>`
	var msg string

	opts := telegram.SendOptions{
		LinkPreview: false,
	}

	if rmsg, err := m.GetReplyMessage(); err != nil {
		return err
	} else if rmsg.Sender != nil || rmsg.SenderChat != nil {
		var replyUserFullName string
		if s := rmsg.Sender; s != nil {
			replyUserFullName = strings.TrimSpace(s.FirstName + " " + s.LastName)
		} else if cs := rmsg.SenderChat; cs != nil {
			replyUserFullName = cs.Title
		}

		msg = fmt.Sprintf(msgTemplate, m.ReplySenderID(), replyUserFullName, m.SenderID(), userFullName, url)
		opts.ReplyID = m.ReplyID()
	} else {
		msg = fmt.Sprintf(msgTemplate, 0, "", m.SenderID(), userFullName, url)
	}

	_, err = m.Respond(msg, opts)
	return orCont(err)
}

func (w *warningTracker) Lock(chatId int64) {
	w.globalLock.Lock()
	lock, exists := w.locks[chatId]
	if !exists {
		lock = &sync.Mutex{}
		w.locks[chatId] = lock
	}
	w.globalLock.Unlock()

	lock.Lock()
}

func (w *warningTracker) Unlock(chatId int64) {
	w.globalLock.Lock()
	lock := w.locks[chatId]
	w.globalLock.Unlock()

	lock.Unlock()
}
