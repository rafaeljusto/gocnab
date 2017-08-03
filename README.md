[![GoDoc](https://godoc.org/github.com/rafaeljusto/gocnab?status.png)](https://godoc.org/github.com/rafaeljusto/gocnab)
[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/rafaeljusto/gocnab/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/rafaeljusto/gocnab)](https://goreportcard.com/report/github.com/rafaeljusto/gocnab)

# gocnab

CNAB (Un)Marshaler will help you to create and/or parse CNAB (Centro Nacional de
Automação Bancária) encoded  files. You can use the struct tags to define the
position of the field in the CNAB files. It will fill with zeros the unused
space when it is a number, or with a space when it is a text.

## Marshal Example

```go
package main

import(
  "fmt"

  "github.com/rafaeljusto/gocnab"
)

type example struct {
  FieldA int         `cnab:"0,20"`
  FieldB string      `cnab:"20,50"`
  FieldC float64     `cnab:"50,60"`
  FieldD uint        `cnab:"60,70"`
  FieldE bool        `cnab:"70,71"`
  FieldF bool        `cnab:"71,80"`
}

func main() {
  e := example{
    FieldA: 123,
    FieldB: "This is a text",
    FieldC: 50.30,
    FieldD: 445,
    FieldE: true,
    FieldF: false,
  }

  data, err := gocnab.Marshal400(e)
  if err != nil {
    fmt.Println(err)
    return
  }

  fmt.Println(data)
}
```
