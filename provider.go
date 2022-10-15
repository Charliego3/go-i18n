package i18n

import (
	"golang.org/x/text/language"
	"net/http"
)

func parseLanguage(val string) language.Tag {
	if len(val) == 0 {
		return language.Tag{}
	}

	tag, err := language.Parse(val)
	if err != nil {
		return language.Tag{}
	}
	return tag
}

func HeaderProvider(_ string, r *http.Request) language.Tag {
	tags, qs, err := language.ParseAcceptLanguage(r.Header.Get(acceptLanguage))
	if err != nil {
		return language.Tag{}
	}

	for i := 0; i < len(tags); {
		if _, ok := g.localizes[tags[i]]; ok {
			i++
			continue
		}

		tags = append(tags[:i], tags[i+1:]...)
		qs = append(qs[:i], qs[i+1:]...)
	}

	if len(tags) == 0 {
		return language.Tag{}
	}

	w := float32(0.0)
	idx := 0
	for i, q := range qs {
		if q > w {
			w = q
			idx = i
		}
	}

	return tags[idx]
}

func CookieProvider(key string, r *http.Request) language.Tag {
	val, err := r.Cookie(key)
	if err != nil {
		return language.Tag{}
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
