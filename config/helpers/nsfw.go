package helpers

import (
	"fmt"
	"regexp"
	"strings"

	"main/config"
	"main/database"
)

func WildcardToRegex(w string) string {
	escaped := regexp.QuoteMeta(w)
	escaped = strings.ReplaceAll(escaped, `\*\*`, `(?s:.*)`)
	escaped = strings.ReplaceAll(escaped, `\*`, `.*`)
	escaped = strings.ReplaceAll(escaped, `\?`, `.`)
	return `(?i)` + escaped
}

func UpdateNSFWRegexCache() error {
	words, err := database.GetNSFWWords()
	if err != nil {
		return err
	}

	var patterns []*regexp.Regexp
	for _, word := range words {
		pattern := WildcardToRegex(word)
		re, err := regexp.Compile(pattern)
		if err == nil {
			patterns = append(patterns, re)
		}
	}

	config.Cache.Store("nsfw_regex", patterns)
	return nil
}

func MatchNSFWText(text string) (bool, string) {
	val, ok := config.Cache.Load("nsfw_regex")
	if !ok {
		_ = UpdateNSFWRegexCache()
		val, ok = config.Cache.Load("nsfw_regex")
		if !ok {
			return false, text
		}
	}

	regexList, ok := val.([]*regexp.Regexp)
	if !ok {
		return false, text
	}

	matched := false
	updated := text

	for _, re := range regexList {
		if re.MatchString(updated) {
			matched = true
			updated = re.ReplaceAllString(updated, "****")
		}
	}

	/*if !matched {
		var err error
		matched, err = IsProfanity(updated)
		if err != nil {
			fmt.Println("IsProfanity Error:", err.Error())
		}

		if matched {
			return true, ""
		}
	}*/

	return matched, updated
}
