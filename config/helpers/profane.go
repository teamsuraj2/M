package helpers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Match struct {
	Match string `json:"match"`
}

type Profanity struct {
	Matches []Match `json:"matches"`
}

type APIResponse struct {
	Status    string    `json:"status"`
	Error     APIError  `json:"error"`
	Profanity Profanity `json:"profanity"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var apiKeys = []map[string]string{
	{"api_user": "1012651687", "api_secret": "kYZmgwUThAGz4yZseukCZTkhNtYp3xsA"},
}

func M(msg string) {
	log.Println("Sightengine Error:", msg)
}

func MatchProfanity(text string) (string, bool) {
	for _, creds := range apiKeys {
		form := url.Values{}
		form.Add("text", text)
		form.Add("mode", "rules")
		form.Add("lang", "en")
		form.Add("categories", "profanity")
		form.Add("api_user", creds["api_user"])
		form.Add("api_secret", creds["api_secret"])

		resp, err := http.Post(
			"https://api.sightengine.com/1.0/text/check.json",
			"application/x-www-form-urlencoded",
			bytes.NewBufferString(form.Encode()),
		)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		var result APIResponse
		json.Unmarshal(body, &result)

		if result.Status != "success" {
			if result.Error.Code == 32 {
				continue
			}
			M(result.Error.Message)
			return text, false
		}

		censored := text
		if len(result.Profanity.Matches) == 0 {
			return text, false
		}
		for _, match := range result.Profanity.Matches {
			censored = strings.ReplaceAll(censored, match.Match, "***")
		}
		return censored, true
	}
	M("All API keys failed or rate-limited")
	return text, false
}
