package modules

import (
	"github.com/amarnathcjd/gogram/telegram"
)

type HelpModule struct {
	Callback string
	Help     string
}

type DeferredHandler struct {
	Pattern any
	Handler any
	Filters []telegram.Filter
}

var Continue error = nil
var (
	ModulesHelp = make(map[string]*HelpModule, 0)
	handler     = make([]DeferredHandler, 0)
)

func LoadMods(c *telegram.Client) {
	c.UpdatesGetState()
        c.On("command:biolink", setBioMode)
	c.On("command:setlonglimit", SetLongLimitHandler)
	c.On("command:setlongmode", SetLongModeHandler)
	c.On("command:echo", EcoHandler)
	c.On("command:ping", pingHandler)
	c.On("command:reload", ReloadHandler)
	c.On("command:start", start)
	c.On("command:stats", stats)
	c.On("command:help", help)
	c.On("command:nolinks", NoLinksCmd)
	c.On("command:allowlink", AllowHostCmd)
	c.On("command:allowhost", AllowHostCmd)
	c.On("command:removelink", RemoveHostCmd)
	c.On("command:listlinks", ListAllowedHosts)

	c.On("message", deleteLongMessage)
	c.On("edit", deleteEditedMessage)
	c.On("message", deleteLinkMessage)
	c.On("message", deleteUserMsgIfBio)
	c.On("participant", botAddded)
	c.On("callback:close", close)
	c.On("callback:help", helpCB)
	c.On("callback:start_callback", startCB)

	for _, h := range handler {
		c.On(h.Pattern, h.Handler, h.Filters...)
	}
}

func AddHelp(name, callback, help string, filters ...telegram.Filter) {
	handler = append(handler, DeferredHandler{
		Pattern: "callback:" + callback,
		Handler: helpModuleCB,
		Filters: filters,
	})
	ModulesHelp[name] = &HelpModule{
		Callback: callback,
		Help:     help,
	}
}

func GetHelp(callback string) string {
	for _, data := range ModulesHelp {
		if data.Callback == callback {
			return data.Help
		}
	}
	return ""
}

func orCont(err error) error {
	if err != nil {
		return err
	}
	return telegram.EndGroup
}
