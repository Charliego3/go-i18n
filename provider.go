package i18n

import (
	"golang.org/x/text/language"
	"net/http"
)

func parseLanguage(val string) language.Tag {
	if len(val) == 0 {
		return language.Und
	}

	tag, err := language.Parse(val)
	if err != nil {
		return language.Und
	}
	return tag
}

func ParseFromHeader(val string) language.Tag {
	tags, _, err := language.ParseAcceptLanguage(val)
	if err != nil {
		return language.Und
	}

	idx := -1
	for i, tag := range tags {
		if _, ok := g.localizes[tag]; ok {
			idx = i
			break
		}
	}

	if idx < 0 {
		return language.Und
	}
	return tags[idx]
}

func HeaderProvider(_ string, r *http.Request) language.Tag {
	return ParseFromHeader(r.Header.Get("Accept-Language"))
}

func CookieProvider(key string, r *http.Request) language.Tag {
	val, err := r.Cookie(key)
	if err != nil {
		return language.Und
	}
	return parseLanguage(val.Value)
}

func QueryProvider(key string, r *http.Request) language.Tag {
	return parseLanguage(r.URL.Query().Get(key))
}

func FormProvider(key string, r *http.Request) language.Tag {
	return parseLanguage(r.FormValue(key))
}

func PostFormProvider(key string, r *http.Request) language.Tag {
	return parseLanguage(r.PostFormValue(key))
}
