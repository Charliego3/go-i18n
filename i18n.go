package i18n

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
	"net/http"
)

type Localized = i18n.LocalizeConfig
type UnmarshalFunc = i18n.UnmarshalFunc
type LanguageProvider func(string, *http.Request) language.Tag
type Message interface {
	~string | Localized | ~*Localized
}

type I18n struct {
	bundle      *i18n.Bundle
	defLanguage language.Tag
	localizes   map[language.Tag]*i18n.Localizer
	providers   []LanguageProvider
	loaders     []Loader
	languageKey string
}

func (g *I18n) setDefaultLang(lang language.Tag) {
	g.defLanguage = lang
	g.bundle = i18n.NewBundle(lang)
}

func (g *I18n) addLoader(loader Loader) {
	result, err := loader.Load()
	if err != nil {
		panic(err)
	}

	for format, fn := range result.Funcs {
		g.registerUnmarshalFunc(format, fn)
	}

	for _, entry := range result.Entries {
		g.setLocalizer(entry.Tag)
		g.mastParseMessageFileBytes(entry.Bytes, entry.Name)
	}
}

func (g *I18n) tr(ctx context.Context, p any) (string, error) {
	lang := g.defLanguage
	if ctx != nil {
		val := ctx.Value(g.languageKey)
		if l, ok := val.(language.Tag); ok {
			lang = l
		}
	}

	lr := g.getLocalizer(lang)
	var config *i18n.LocalizeConfig
	switch t := p.(type) {
	case string:
		config = &i18n.LocalizeConfig{MessageID: t}
	case i18n.LocalizeConfig:
		config = &t
	case *i18n.LocalizeConfig:
		config = t
	default:
		return "", fmt.Errorf("unsupported param %T", p)
	}

	translated, err := lr.Localize(config)
	if err != nil || len(translated) == 0 {
		return config.MessageID, err
	}
	return translated, nil
}

func (g *I18n) getLanguage(r *http.Request) language.Tag {
	for _, provider := range g.providers {
		tag := provider(g.languageKey, r)
		if !tag.IsRoot() {
			return tag
		}
	}
	return g.defLanguage
}

func getUnmarshaler(format string) UnmarshalFunc {
	switch format {
	case "json":
		return json.Unmarshal
	case "toml":
		return toml.Unmarshal
	case "yaml":
		return yaml.Unmarshal
	default:
		return nil
	}
}

func (g *I18n) registerUnmarshalFunc(format string, unmarshalFunc UnmarshalFunc) {
	g.bundle.RegisterUnmarshalFunc(format, unmarshalFunc)
}

func (g *I18n) mastParseMessageFileBytes(buf []byte, path string) {
	g.bundle.MustParseMessageFileBytes(buf, path)
}

func (g *I18n) setLocalizer(lang language.Tag) {
	if _, ok := g.localizes[lang]; ok {
		return
	}

	if g.localizes == nil {
		g.localizes = make(map[language.Tag]*i18n.Localizer)
	}

	langs := []string{lang.String()}
	if lang != g.defLanguage {
		langs = append(langs, g.defLanguage.String())
	}

	g.localizes[lang] = i18n.NewLocalizer(g.bundle, langs...)
}

func (g *I18n) getLocalizer(lang language.Tag) *i18n.Localizer {
	if lr, ok := g.localizes[lang]; ok {
		return lr
	}

	return g.localizes[g.defLanguage]
}

var g *I18n

func Initialize(opts ...Option) *I18n {
	if g != nil {
		return g
	}

	g = &I18n{}
	for _, opt := range opts {
		opt(g)
	}

	if g.defLanguage.IsRoot() {
		g.setDefaultLang(language.English)
	} else {
		g.setDefaultLang(g.defLanguage)
	}

	for _, loader := range g.loaders {
		g.addLoader(loader)
	}

	if len(g.providers) == 0 {
		g.providers = []LanguageProvider{
			HeaderProvider,
			CookieProvider,
			QueryProvider,
			FormProvider,
			PostFormProvider,
		}
	}

	if len(g.languageKey) == 0 {
		g.languageKey = "accept-language"
	}
	g.setLocalizer(g.defLanguage)
	return g
}

// Handler returns http.Handler. It can be using a middleware...
func (g *I18n) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tag := g.getLanguage(r)
		r = r.WithContext(context.WithValue(r.Context(), g.languageKey, tag))
		next.ServeHTTP(w, r)
	})
}

// Tr translate messageId to target language
//
// Example:
//
//	Tr(ctx, "hello")
//	Tr(ctx, &Localized{
//		Message: "HelloName",
//		TemplateData: map[string]string{
//			"Name": "I18n",
//		},
//	})
func Tr[T Message](ctx context.Context, message T) (string, error) {
	if g == nil {
		panic("i18n uninitialized, using i18n.Initialize(opts...) to init")
	}
	return g.tr(ctx, message)
}

// MustTr called Tr but ignore error
func MustTr[T Message](ctx context.Context, message T) string {
	translated, _ := Tr(ctx, message)
	return translated
}
