package i18n

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

type LanguageKey struct{}
type Localized = i18n.LocalizeConfig
type UnmarshalFunc = i18n.UnmarshalFunc
type LanguageProvider[T, U any] func(source T, key U) language.Tag
type Message interface {
	~string | Localized | ~*Localized
}

type I18n struct {
	bundle      *i18n.Bundle
	defaultlang language.Tag
	localizes   map[language.Tag]*i18n.Localizer
	loaders     []Loader
	languageKey any
}

func (g *I18n) setDefaultLang(lang language.Tag) {
	g.defaultlang = lang
	g.bundle = i18n.NewBundle(lang)
}

func (g *I18n) addLoader(loader Loader) {
	result, err := loader.Load()
	if err != nil {
		panic(err)
	}

	for format, fn := range result.Funcs {
		g.bundle.RegisterUnmarshalFunc(format, fn)
	}

	for _, entry := range result.Entries {
		g.setLocalizer(entry.Tag)
		g.bundle.MustParseMessageFileBytes(entry.Bytes, entry.Name)
	}
}

func (g *I18n) tr(ctx context.Context, p any) (string, error) {
	value := ctx.Value(LanguageKey{})
	tag, ok := value.(language.Tag)
	if !ok {
		tag = g.defaultlang
	}

	lr := g.getLocalizer(tag)
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

func (g *I18n) setLocalizer(lang language.Tag) {
	if _, ok := g.localizes[lang]; ok {
		return
	}

	if g.localizes == nil {
		g.localizes = make(map[language.Tag]*i18n.Localizer)
	}

	langs := []string{lang.String()}
	if lang != g.defaultlang {
		langs = append(langs, g.defaultlang.String())
	}

	g.localizes[lang] = i18n.NewLocalizer(g.bundle, langs...)
}

func (g *I18n) getLocalizer(lang language.Tag) *i18n.Localizer {
	if lr, ok := g.localizes[lang]; ok {
		return lr
	}

	return g.localizes[g.defaultlang]
}

var (
	g *I18n

	languageProvider any
)

func Initialize(opts ...Option) *I18n {
	if g != nil {
		return g
	}

	g = new(I18n)
	for _, opt := range opts {
		opt(g)
	}

	if g.defaultlang.IsRoot() {
		g.setDefaultLang(language.English)
	} else {
		g.setDefaultLang(g.defaultlang)
	}

	for _, loader := range g.loaders {
		g.addLoader(loader)
	}

	if g.languageKey == nil {
		g.languageKey = "Accept-Language"
	}

	g.setLocalizer(g.defaultlang)
	return g
}

// Handler returns http.Handler. It can be using a middleware...
func (g *I18n) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tag := g.defaultlang
		if p, ok := languageProvider.(LanguageProvider[*http.Request, string]); ok {
			tag = p(r, g.languageKey.(string))
		}
		r = r.WithContext(context.WithValue(r.Context(), LanguageKey{}, tag))
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
