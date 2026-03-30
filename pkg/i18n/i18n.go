package i18n

import (
	"sync"
)

var (
	translator *Translator
	once       sync.Once
)

type Translator struct {
	lang string
	t    map[string]string
}

func NewTranslator(lang string) *Translator {
	t := &Translator{lang: lang}
	t.loadTranslations()
	return t
}

func GetTranslator(lang string) *Translator {
	once.Do(func() {
		translator = NewTranslator(lang)
	})
	if translator.lang != lang {
		translator.SetLang(lang)
	}
	return translator
}

func (t *Translator) T(key string) string {
	if val, ok := t.t[key]; ok {
		return val
	}
	return key
}

func (t *Translator) SetLang(lang string) {
	t.lang = lang
	t.loadTranslations()
}

func (t *Translator) loadTranslations() {
	switch t.lang {
	case "en":
		t.t = en
	default:
		t.t = zh
	}
}
