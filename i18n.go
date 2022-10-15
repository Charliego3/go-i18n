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

const acceptLanguage = "Accept-Language"

type LocalizeConfig = i18n.LocalizeConfig
type UnmarshalFunc = i18n.UnmarshalFunc
type LanguageProvider func(string, *http.Request) language.Tag

type options struct {
	bundle      *i18n.Bundle
	defLanguage language.Tag
	localizes   map[language.Tag]*i18n.Localizer
	providers   []LanguageProvider
	loaders     []Loader
	languageKey string
}

func (o *options) setDefaultLang(lang language.Tag) {
	o.defLanguage = lang
	o.bundle = i18n.NewBundle(lang)
}

func (o *options) AddLoader(loader Loader) {
	err := loader.ParseMessage(o)
	if err != nil {
		panic(err)
	}
	o.SetLocalizer(o.defLanguage)
}

func (o *options) tr(ctx context.Context, p any) (string, error) {
	lang := o.defLanguage
	if ctx != nil {
		val := ctx.Value(acceptLanguage)
		if l, ok := val.(language.Tag); ok {
			lang = l
		}
	}

	lr := o.getLocalizer(lang)
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

func (o *options) getLanguage(r *http.Request) language.Tag {
	for _, provider := range o.providers {
		tag := provider(o.languageKey, r)
		if !tag.IsRoot() {
			return tag
		}
	}
	return o.defLanguage
}

func (o *options) registerUnmarshalFunc(format string) {
	var fn UnmarshalFunc
	switch format {
	case "json":
		fn = json.Unmarshal
	case "toml":
		fn = toml.Unmarshal
	case "yaml":
		fn = yaml.Unmarshal
	}

	if fn != nil {
		o.RegisterUnmarshalFunc(format, fn)
	}
}

func (o *options) RegisterUnmarshalFunc(format string, unmarshalFunc UnmarshalFunc) {
	o.bundle.RegisterUnmarshalFunc(format, unmarshalFunc)
}

func (o *options) MastParseMessageFileBytes(buf []byte, path string) {
	o.bundle.MustParseMessageFileBytes(buf, path)
}

func (o *options) SetLocalizer(lang language.Tag) {
	if _, ok := o.localizes[lang]; ok {
		return
	}

	if o.localizes == nil {
		o.localizes = make(map[language.Tag]*i18n.Localizer)
	}

	langs := []string{lang.String()}
	if lang != o.defLanguage {
		langs = append(langs, o.defLanguage.String())
	}

	o.localizes[lang] = i18n.NewLocalizer(o.bundle, langs...)
}

func (o *options) getLocalizer(lang language.Tag) *i18n.Localizer {
	if lr, ok := o.localizes[lang]; ok {
		return lr
	}

	return o.localizes[o.defLanguage]
}

var gopt *options

// Localize initialize i18n...
func Localize(next http.Handler, opts ...Option) http.Handler {
	gopt = &options{}
	for _, opt := range opts {
		opt(gopt)
	}

	if gopt.defLanguage.IsRoot() {
		gopt.setDefaultLang(language.English)
	}

	for _, loader := range gopt.loaders {
		gopt.AddLoader(loader)
	}

	if len(gopt.providers) == 0 {
		gopt.providers = []LanguageProvider{
			HeaderProvider,
			CookieProvider,
			QueryProvider,
			FormProvider,
			PostFormProvider,
		}
	}

	if len(gopt.languageKey) == 0 {
		gopt.languageKey = "lang"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tag := gopt.getLanguage(r)
		r = r.WithContext(context.WithValue(r.Context(), acceptLanguage, tag))
		next.ServeHTTP(w, r)
	})
}

// Tr translate messageId to target language
//
// Example:
//
//	Tr("hello")
//	Tr(LocalizeConfig{
//		MessageID: "HelloName",
//		TemplateData: map[string]string{
//			"Name": "I18n",
//		},
//	})
func Tr[T string | LocalizeConfig | *LocalizeConfig](ctx context.Context, messageId T) (string, error) {
	return gopt.tr(ctx, messageId)
}

// MustTr called Tr but ignore error
func MustTr[T string | LocalizeConfig | *LocalizeConfig](ctx context.Context, messageId T) string {
	message, _ := Tr(ctx, messageId)
	return message
}
