package i18n

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

var _ I18n = (*i18nImpl)(nil)

type i18nImpl struct {
	bundle      *i18n.Bundle
	loader      Loader
	language    language.Tag
	localizes   map[string]*i18n.Localizer
	ctx         context.Context
	langHandler LangHandler
}

func (i *i18nImpl) RegisterLang(lang language.Tag) {
	i.language = lang
	i.bundle = i18n.NewBundle(lang)
}

func (i *i18nImpl) AddLoader(loader Loader) {
	unmarshalers := loader.GetUnmarshalers()
	for format, fn := range unmarshalers {
		i.bundle.RegisterUnmarshalFunc(format, fn)
	}

	err := loader.Walk(func(path, ext string, lang language.Tag) error {
		if _, ok := unmarshalers[ext]; !ok {
			i.registerUnmarshalFunc(ext)
		}
		i.setLocalizer(lang)
		buf, err := loader.LoadMessage(path)
		if err != nil {
			return err
		}

		i.bundle.MustParseMessageFileBytes(buf, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	i.setLocalizer(i.language)
}

func (i *i18nImpl) GetMessage(p interface{}) (string, error) {
	lang := i.getLangHandler().GetLang(i.ctx, i.language.String())
	lr := i.getLocalizer(lang)

	var config *i18n.LocalizeConfig

	switch t := p.(type) {
	case string:
		config = &i18n.LocalizeConfig{
			MessageID: t,
		}
	case *TemplateConfig:
		config = t.GetConfig()
	case *i18n.LocalizeConfig:
		config = t
	default:
		return config.MessageID, fmt.Errorf("unsupported param %T", p)
	}

	translated, err := lr.Localize(config)
	if err != nil {
		return config.MessageID, err
	}
	if len(translated) == 0 {
		return config.MessageID, nil
	}
	return translated, nil
}

func (i *i18nImpl) MustGetMessage(p interface{}) string {
	message, _ := i.GetMessage(p)
	return message
}

func (i *i18nImpl) SetLangHandler(handler LangHandler) {
	i.langHandler = handler
}

func (i *i18nImpl) SetCurrentContext(ctx context.Context) {
	i.ctx = ctx
}

func (i *i18nImpl) Apply(I18n) {}

func (i *i18nImpl) getLangHandler() LangHandler {
	if i.langHandler != nil {
		return i.langHandler
	}
	i.langHandler = &defaultLangHandler{}
	return i.langHandler
}

func (i *i18nImpl) registerUnmarshalFunc(ext string) {
	switch ext {
	case "json":
		i.bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	case "toml":
		i.bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	case "yaml":
		i.bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	}
}

func (i *i18nImpl) setLocalizer(lang language.Tag) {
	if i.localizes == nil {
		i.localizes = make(map[string]*i18n.Localizer)
	}

	if _, ok := i.localizes[lang.String()]; ok {
		return
	}

	langs := []string{lang.String()}
	if lang != i.language {
		langs = append(langs, i.language.String())
	}

	i.localizes[lang.String()] = i18n.NewLocalizer(i.bundle, langs...)
}

func (i *i18nImpl) getLocalizer(lang string) *i18n.Localizer {
	if lr, ok := i.localizes[lang]; ok {
		return lr
	}

	return i.localizes[i.language.String()]
}

type defaultLangHandler struct{}

func (g *defaultLangHandler) GetLang(context context.Context, defaultLang string) string {
	var ctx *gin.Context
	if context == nil {
		return defaultLang
	}

	ctx = context.(*gin.Context)
	if ctx.Request == nil {
		return defaultLang
	}

	lan := getLan(func() string {
		return ctx.GetHeader("Accept-Language")
	}, func() string {
		return ctx.Query("lang")
	}, func() string {
		return ctx.Request.FormValue("lang")
	}, func() string {
		return ctx.Request.PostFormValue("lang")
	})

	if len(lan) > 0 {
		return lan
	}

	return defaultLang
}

func getLan(fns ...func() string) string {
	for _, fn := range fns {
		lan := fn()
		if len(lan) > 0 {
			return lan
		}
	}
	return ""
}
