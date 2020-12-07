package common

import (
	"encoding/json"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var bundle *i18n.Bundle

func InitializeTranslations() {
	bundle = i18n.NewBundle(language.French)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.MustLoadMessageFile("./translations/en.json")
	bundle.MustLoadMessageFile("./translations/fr.json")
	bundle.MustLoadMessageFile("./translations/ca.json")
}

func Translate(word, language string) string {
	loc := i18n.NewLocalizer(bundle, language)
	translated, err := loc.Localize(&i18n.LocalizeConfig{
		MessageID: word,
	})
	if err == nil {
		return translated
	} else {
		loc = i18n.NewLocalizer(bundle, "fr")
		return loc.MustLocalize(&i18n.LocalizeConfig{
			MessageID: word,
		})
	}
}
