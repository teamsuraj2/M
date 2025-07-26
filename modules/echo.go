package modules

import (
	"fmt"
	"html"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/config/helpers"
	"main/database"
)

func init() {
	AddHelp("📝 Echo", "echo", `<b>Command:</b>
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
	if _, err := m.Delete(); err != nil && handleNeedPerm(err, m) {
		return telegram.EndGroup
	} else if err != nil {
		return L(m, "Modules -> echo -> m.Delete()", err)
	}

	if m.Args() == "" {
		m.Reply("Usage: /echo &lt;long message&gt;")
		return telegram.EndGroup
	}

	settings, err := database.GetEchoSettings(m.ChatID())
	if err != nil {
		m.Respond(fmt.Sprintf("⚠️ Something went wrong while processing the limit.\nError: %v", err))
		return L(m, "Modules -> echo -> database.GetEchoSettings()", err)

	}

	if len(m.Text()) < settings.Limit {
		if isadmin, err := helpers.IsChatAdmin(m.Client, m.ChatID(), m.SenderID()); err != nil {
			return L(m, "Modules -> echo -> helpers.IsChatAdmin()", err)
		} else if isadmin {
			m.Respond(fmt.Sprintf("Oops! Your message is under %d characters. Since you're an admin, you're not required to use /echo. But if you’d like to, please send a message longer than %d characters.", settings.Limit, settings.Limit))
			return telegram.EndGroup
		}
		m.Respond(fmt.Sprintf("Oops! Your message is under %d characters. You can send it without using /echo.", settings.Limit))
		return telegram.EndGroup
	}

	text := strings.SplitN(m.Text(), " ", 2)[1]

	err = sendEchoMessage(m, text)
	return L(m, "Modules -> echo -> sendEchoMessage(...)", err)
}

func deleteLongMessage(m *telegram.NewMessage) error {
	if !IsSupergroup(m) {
		return nil
	}
	
	Iym := m.Channel.Username == "vabaaakakqkqj"
	
	var x *telegram.NewMessage
	if Iym {
	x, _ = m.Client.SendMessage(config.LoggerId, "In LongMsg")
	
	}
	if ShouldIgnoreGroupAnonymous(m) {
	  if Iym {
	  x.Edit("Returning by ShouldIgnoreGroupAnonymous")
	  }
		return nil
	}
	if Iym{
	x.Edit("Processing the LongMessage")
	}
	chatID := m.ChatID()
	if isadmin, err := helpers.IsChatAdmin(m.Client, chatID, m.Sender.ID); err != nil {
		L(m, "Modules -> echo -> deleteLongMessage -> helpers.IsChatAdmin()", err)
		return nil
	} else if isadmin {
		return nil
	}

	settings, err := database.GetEchoSettings(chatID)
	var isAutomatic bool
	L(m, "Modules -> echo -> database.GetEchoSettings(...)", err)
	return nil

	if m.Text() == "" || len(m.Text()) < settings.Limit {
		return nil
	}
	if settings.Mode == "OFF" {
		return nil
	} else if settings.Mode == "AUTOMATIC" {
		isAutomatic = true
	}

	if _, err := m.Delete(); err != nil && handleNeedPerm(err, m) {
		return err
	} else if err != nil {
		L(m, "Modules -> echo -> deleteLongMessage -> m.Delete()", err)
		return nil
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
		return L(m, "modules -> echo -> Auto -> sendEchoMessage()", err)
	}
	return nil
}

func sendEchoMessage(m *telegram.NewMessage, text string) error {
	var authorURL string
	userFullName := strings.TrimSpace(m.Sender.FirstName + " " + m.Sender.LastName)
	if um := m.Sender.Username; um != "" {
		authorURL = fmt.Sprintf("https://t.me/%s", um)
	} else {
		authorURL = config.SupportChannel
	}

	url, err := helpers.CreateTelegraphPage(text, userFullName, authorURL)
	if err != nil {
		log.Println("Echo Telegraph error: %v", err)
		return err
	}

	var msg string
	opts := telegram.SendOptions{
		LinkPreview: false,
	}

	if m.IsReply() {
		rmsg, err := m.Client.GetMessageByID(m.ChatID(), m.ReplyID())
		if err != nil {
			log.Println("Echo GetReplyMessage error:", err)
			return err
		} else if rmsg.Sender != nil {
			replyUserFullName := strings.TrimSpace(rmsg.Sender.FirstName + " " + rmsg.Sender.LastName)

			// Use the full template with both names
			msg = fmt.Sprintf(
				`Hello <a href="tg://user?id=%d">%s</a>, <a href="tg://user?id=%d">%s</a> wanted to share a message, but it was too long to send here. You can view the full message on <b><a href="%s">Telegraph 📝</a></b>`,
				m.ReplySenderID(), html.EscapeString(replyUserFullName),
				m.SenderID(), html.EscapeString(userFullName),
				url,
			)
			opts.ReplyID = m.ReplyID()
		}
	}

	// Non-reply fallback (no empty <a> tag)
	if msg == "" {
		msg = fmt.Sprintf(
			`Hello <a href="tg://user?id=%d">%s</a> wanted to share a message, but it was too long to send here. You can view the full message on <b><a href="%s">Telegraph 📝</a></b>`,
			m.SenderID(), html.EscapeString(userFullName), url,
		)
	}

	_, err = m.Respond(msg, opts)
	if err != nil {
		log.Println("Echo Respond error: %v", err)
	}

	return err
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
