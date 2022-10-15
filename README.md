# go-i18n

[![GitHub Workflow Status (branch)](https://img.shields.io/github/workflow/status/Charliego93/go-i18n/Go/main?logo=github)](https://github.com/Charliego93/go-i18n/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Charliego93/go-i18n)](https://goreportcard.com/report/github.com/Charliego93/go-i18n)
[![Coveralls](https://img.shields.io/coveralls/github/Charliego93/go-i18n?logo=coveralls&color=8345D5)](https://coveralls.io/github/Charliego93/go-i18n?branch=main)
[![GitHub tag (latest SemVer pre-release)](https://img.shields.io/github/v/tag/Charliego93/go-i18n?include_prereleases)](https://github.com/Charliego93/go-i18n/tags)
[![GitHub](https://img.shields.io/github/license/Charliego93/go-i18n?color=D96BA2)](https://github.com/Charliego93/go-i18n/blob/main/LICENSE)
[![GoDoc](https://godoc.org/github.com/Charliego93/go-i18n?status.svg)](https://godoc.org/github.com/Charliego93/go-i18n)

Provides simplicity and ease of use, no specific framework restrictions, easy access to any framework based on `http.Handler`

# Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Customize Loader](#customize-loader)
- [License](#license)

## Installation

```shell
go get github.com/Charliego93/go-i18n
```

## Usage

```go
package main

import (
   "embed"
   "fmt"
   "github.com/Charliego93/go-i18n/v2"
   "github.com/gin-gonic/gin"
   "golang.org/x/text/language"
   "net/http"
)

//go:embed examples/lan2/*
var langFS embed.FS

func main() {
   engine := gin.New()

   // curl -H "Accept-Language: en" 'http://127.0.0.1:9090/Hello'  returns "hello"
   // curl -H "Accept-Language: uk" 'http://127.0.0.1:9090/Hello'  returns "Бонгу"
   // curl 'http://127.0.0.1:9090/Hello?lang=en'  returns "hello"
   // curl 'http://127.0.0.1:9090/Hello?lang=uk'  returns "Бонгу"
   engine.GET("/:messageId", func(ctx *gin.Context) {
      ctx.String(http.StatusOK, i18n.MustTr(ctx.Request.Context(), ctx.Param("messageId")))
   })

   // curl -H "Accept-Language: en" 'http://127.0.0.1:9090/HelloName/I18n'  returns "hello I18n"
   // curl -H "Accept-Language: uk" 'http://127.0.0.1:9090/HelloName/I18n'  returns "Бонгу I18n"
   // curl 'http://127.0.0.1:9090/HelloName/I18n?lang=en'  returns "hello I18n"
   // curl 'http://127.0.0.1:9090/HelloName/I18n?lang=uk'  returns "Бонгу I18n"
   engine.GET("/:messageId/:name", func(ctx *gin.Context) {
      ctx.String(http.StatusOK, i18n.MustTr(ctx.Request.Context(), &i18n.LocalizeConfig{
         MessageID: ctx.Param("messageId"),
         TemplateData: map[string]string{
            "Name": ctx.Param("name"),
         },
      }))
   })

   // Use multi loader provider
   // Built-in load from file and load from fs.FS
   // i18n.Initialize(i18n.NewLoaderWithFS(langFS), i18n.NewLoaderWithPath("./examples/lan1"))))
   g := i18n.Initialize(i18n.NewLoaderWithPath("./examples/simple"))
   if err := http.ListenAndServe(":9090", g.Handler(engine)); err != nil {
      panic(err)
   }
}
```

## Customize Loader

You can implement your own `Loader` by yourself, and even pull the language files from anywhere

```go
type Loader interface {
    Load() (*Result, error)
}

type Result struct {
   Funcs   map[string]UnmarshalFunc
   Entries []Entry
}

type Entry struct {
   Lauguage language.Tag
   Name     string
   Bytes    []byte
}
```

## License

[MIT © Charliego93.](LICENSE)
