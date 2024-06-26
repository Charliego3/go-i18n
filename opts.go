package i18n

import "golang.org/x/text/language"

type Option func(*I18n)

// WithLoader Register the Loader interface to *I18n.bundle
//
// Example:
//
//	//go:embed examples/lan2/*
//	var langFS embed.FS
//	i18n.Handler(http.Handler, i18n.NewLoaderWithPath("language_file_path"))
//	i18n.Handler(http.Handler, i18n.NewLoaderWithFS(langFS, i18n.WithUnmarshal("json", json.Unmarshal)))
func WithLoader(loader Loader) Option {
	return func(o *I18n) {
		o.loaders = append(o.loaders, loader)
	}
}

// WithLanguageProvider get the language from *http.Request,
// default LanguageProvider the order of acquisition is: header(always get the value of Accept-Language) -> cookie -> query -> form -> postForm
// you can use WithLanguageKey change the default lang key
//
// Example:
//
//	loader := i18n.NewLoaderWithPath("language_file_path")
//	i18n.Handler(http.Handler, i18n.WithLoader(loader),
//		i18n.WithProvider(i18n.HeaderProvider)
//	)
func WithProvider[T, U any](provider LanguageProvider[T, U]) Option {
	return func(o *I18n) {
		languageProvider = provider
	}
}

// WithLanguageKey specifies the default language key when obtained from the LanguageProvider
// Except from the Header, there is no limit if you specify LanguageProvider manually
//
// Example:
//
//	i18n.loader :=i18n.NewLoaderWithPath("language_file_path")
//	i18n.Handler(http.Handler, i18n.WithLoader(loader), i18n.WithLanguageKey("default_language_key"))
func WithLanguageKey(key any) Option {
	return func(o *I18n) {
		o.languageKey = key
	}
}

// WithDefaultLanguage specify the default language,
// which is used when it is not available from the LanguageProvider
//
// Example:
//
//	i18n.loader :=i18n.NewLoaderWithPath("language_file_path")
//	i18n.Handler(http.Handler, i18n.WithLoader(loader), i18n.WithDefaultLanguage(language.Chinese))
func WithDefaultLanguage(tag language.Tag) Option {
	return func(o *I18n) {
		o.defaultlang = tag
	}
}

// Loader Options

type LOpt func(*fsLoader)

// WithUnmarshalls register multi format unmarshal func
func WithUnmarshalls(fns map[string]UnmarshalFunc) LOpt {
	return func(l *fsLoader) {
		l.rs.Funcs = fns
	}
}

// WithUnmarshal register single format unmarshal func
func WithUnmarshal(format string, fn UnmarshalFunc) LOpt {
	return func(l *fsLoader) {
		l.rs.Funcs[format] = fn
	}
}
