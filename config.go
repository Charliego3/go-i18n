package i18n

import "github.com/nicksnyder/go-i18n/v2/i18n"

type LocalizeConfig = interface{}

type TemplateConfig struct {
	MessageID    string
	TemplateData map[string]interface{}
	PluralCount  interface{}
}

func (l *TemplateConfig) GetConfig() *i18n.LocalizeConfig {
	return &i18n.LocalizeConfig{
		MessageID:    l.MessageID,
		TemplateData: l.TemplateData,
		PluralCount:  l.PluralCount,
	}
}
