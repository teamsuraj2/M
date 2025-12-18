package modules

import "github.com/amarnathcjd/gogram/telegram"

func closeCB(cq *telegram.CallbackQuery) error {
	cq.Answer("")
	cq.Delete()
	return telegram.EndGroup
}
