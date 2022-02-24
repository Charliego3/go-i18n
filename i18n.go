package i18n

import (
	"context"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
)

type I18n interface {
	Option
	RegisterLang(language.Tag)
	AddLoader(Loader)
	GetMessage(interface{}) (string, error)
	MustGetMessage(interface{}) string
	SetLangHandler(LangHandler)
	SetCurrentContext(context.Context)
}

type Option interface{ Apply(I18n) }
type OptionFunc func(I18n)
type loader struct{ Loader }
type langHandler struct{ LangHandler }
type LangHandlerFunc func(context context.Context, defaultLng string) string
type LangHandler interface {
	GetLang(context context.Context, defaultLang string) string
}

func (l loader) Apply(n I18n)      { n.AddLoader(l) }
func (f OptionFunc) Apply(n I18n)  { f(n) }
func (h langHandler) Apply(n I18n) { n.SetLangHandler(h) }
func (f LangHandlerFunc) GetLang(context context.Context, defaultLang string) string {
	return f(context, defaultLang)
}

func WithLoader(l Loader) Option                 { return loader{l} }
func WithLangHandler(handler LangHandler) Option { return langHandler{handler} }

var instance I18n

func Localize(defaultLang language.Tag, opts ...Option) gin.HandlerFunc {
	if n, ok := opts[0].(I18n); ok {
		instance = n
	} else if instance == nil {
		instance = &i18nImpl{}
	}

	instance.RegisterLang(defaultLang)
	for _, opt := range opts {
		opt.Apply(instance)
	}

	return func(ctx *gin.Context) {
		instance.SetCurrentContext(ctx)
	}
}

func GetMessage(messageId LocalizeConfig) (string, error) {
	return instance.GetMessage(messageId)
}

func MustGetMessage(messageId LocalizeConfig) string {
	return instance.MustGetMessage(messageId)
}
