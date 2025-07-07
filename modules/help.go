package modules

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/amarnathcjd/gogram/telegram"
)

func help(m *telegram.NewMessage) error {
	m.Delete()
	if m.ChatType() != telegram.EntityUser {
		keyboard := telegram.Button.Keyboard(
			telegram.Button.Row(
				telegram.Button.URL("🗒 Command", fmt.Sprintf("https://t.me/%s?start=help", m.Client.Me().Username)),
			),
		)
		_, err := m.Respond(
			"Contact me in PM for help!", telegram.SendOptions{
				ReplyMarkup: keyboard,
			})
		return L(m, "Modules -> help -> pvt-respond", err)
	}

	keyboard := telegram.NewKeyboard()

	var buttons []telegram.KeyboardButton
	keys := make([]string, 0, len(ModulesHelp))
	for name := range ModulesHelp {
		keys = append(keys, name)
	}

	sort.Slice(keys, func(i, j int) bool {
		return stripPrefixEmoji(keys[i]) < stripPrefixEmoji(keys[j])
	})

	for _, name := range keys {
		mod := ModulesHelp[name]
		button := telegram.Button.Data(name, mod.Callback)
		buttons = append(buttons, button)
	}

	keyboard.NewColumn(2, buttons...)
	keyboard.AddRow(telegram.Button.Data("🗑️ Close", "close"))
	helpText := `📚 <b>Bot Command Help</b>

Here you'll find details for all available plugins and features.

👇 Tap the buttons below to view help for each module:`

	_, err := m.Respond(helpText, telegram.SendOptions{
		ReplyMarkup: keyboard.Build(),
	})
	return L(m, "Modules -> help -> grp/respond", err)
}

func helpCB(c *telegram.CallbackQuery) error {
	keyboard := telegram.NewKeyboard()

	var buttons []telegram.KeyboardButton
	keys := make([]string, 0, len(ModulesHelp))
	for name := range ModulesHelp {
		keys = append(keys, name)
	}

	sort.Slice(keys, func(i, j int) bool {
		return stripPrefixEmoji(keys[i]) < stripPrefixEmoji(keys[j])
	})

	for _, name := range keys {
		mod := ModulesHelp[name]
		button := telegram.Button.Data(name, mod.Callback)
		buttons = append(buttons, button)
	}

	keyboard.NewColumn(2, buttons...)
	keyboard.AddRow(telegram.Button.Data("⬅️ Back", "start_callback"))

	helpText := `📚 <b>Bot Command Help</b>

Here you'll find details for all available plugins and features.

👇 Tap the buttons below to view help for each module:`

	_, err := c.Edit(helpText, &telegram.SendOptions{
		ReplyMarkup: keyboard.Build(),
	})
	return err
}

func helpModuleCB(c *telegram.CallbackQuery) error {
	var helpText string
	for _, module := range ModulesHelp {
		if module.Callback == string(c.Data) {
			helpText = module.Help
			break
		}
	}

	if helpText == "" {
		helpText = "❌ No help found for this module."
	}

	keyboard := telegram.Button.Keyboard(
		telegram.Button.Row(
			telegram.Button.Data("⬅️ Back", "help"),
		),
	)

	_, err := c.Edit(helpText, &telegram.SendOptions{
		ReplyMarkup: keyboard,
	})
	return err
}

func stripPrefixEmoji(s string) string {
	for i, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return strings.ToLower(s[i:])
		}
	}
	return s
}