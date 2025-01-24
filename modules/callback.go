package modules

import "github.com/amarnathcjd/gogram/telegram"

func close(cq *telegram.CallbackQuery) error {
	cq.Answer("")
	cq.Delete()
	return telegram.EndGroup
}
