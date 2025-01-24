package helpers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

func IsProfanity(message string) (bool, error) {
	if len(strings.Fields(message)) == 1 {
		message += " "
	}

	payload := map[string]string{"message": message}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	resp, err := http.Post("https://vector.profanity.dev", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result struct {
		IsProfanity bool    `json:"isProfanity"`
		Score       float64 `json:"score"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		return false, errors.New("non-200 response from server")
	}

	return result.IsProfanity, nil
}
