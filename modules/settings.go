package modules 


func HandleWebAppCommand(m *telegram.NewMessage) error {
		webButton := telegram.Button.WebView("🌐 Open WebApp", "https://gotgboy-571497f84322.herokuapp.com/")
		markup := telegram.Button.Keyboard(
			telegram.Button.Row(webButton),
		)

		m.Respond("Click the button below to open the WebApp:", telegram.SendOptions{
			ReplyMarkup: markup,
		})
	return nil
}