[![GoDoc](https://godoc.org/github.com/rafaeljusto/gocnab?status.png)](https://godoc.org/github.com/rafaeljusto/gocnab)
[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/rafaeljusto/gocnab/master/LICENSE)
[![Build Status](https://travis-ci.org/rafaeljusto/gocnab.svg?branch=master)](https://travis-ci.org/rafaeljusto/gocnab)
[![Coverage Status](https://coveralls.io/repos/github/rafaeljusto/gocnab/badge.svg?branch=master)](https://coveralls.io/github/rafaeljusto/gocnab?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/rafaeljusto/gocnab)](https://goreportcard.com/report/github.com/rafaeljusto/gocnab)
[![codebeat badge](https://codebeat.co/badges/b3a4c784-49db-4e3f-81f7-c35f4e35f70a)](https://codebeat.co/projects/github-com-rafaeljusto-gocnab-master)

![toglacier](https://raw.githubusercontent.com/rafaeljusto/gocnab/master/gocnab.png)

# gocnab

CNAB (Un)Marshaler is an encoder for the brazilian banks' protocol that will
help you to create and/or parse CNAB (Centro Nacional de Automação Bancária)
encoded files. You can use the struct tags to define the position of the field
in the CNAB files `[begin,end)`. It supports basic types `string`, `bool`
(represented by `1` or `0`), `int`, `int8`, `int16`, `int32`, `int64`, `uint`,
`uint8`, `uint16`, `uint23`, `uint64`, `float32` and `float64` (decimal
separator removed). And for custom types it is possible to implement
`gocnab.Marshaler`, `gocnab.Unmarshaler`, `encoding.TextMarshaler` and
`encoding.TextUnmarshaler` to make full use of this library.

## Install

```
go get -u github.com/rafaeljusto/gocnab
```

## Usage

```go
package main

import (
	"fmt"
	"reflect"

	"github.com/rafaeljusto/gocnab"
)

type example struct {
	FieldA int     `cnab:"0,20"`
	FieldB string  `cnab:"20,50"`
	FieldC float64 `cnab:"50,60"`
	FieldD uint    `cnab:"60,70"`
	FieldE bool    `cnab:"70,71"`
}

func main() {
	e1 := example{
		FieldA: 123,
		FieldB: "This is a text",
		FieldC: 50.30,
		FieldD: 445,
		FieldE: true,
	}

	data, err := gocnab.Marshal400(e1)
	if err != nil {
		fmt.Println(err)
		return
	}

	var e2 example
	if err = gocnab.Unmarshal(data, &e2); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(reflect.DeepEqual(e1, e2))
}
```
