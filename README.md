[![GoDoc](https://godoc.org/github.com/rafaeljusto/gocnab?status.png)](https://godoc.org/github.com/rafaeljusto/gocnab)
[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/rafaeljusto/gocnab/master/LICENSE)
[![Build Status](https://travis-ci.org/rafaeljusto/gocnab.svg?branch=master)](https://travis-ci.org/rafaeljusto/gocnab)
[![Coverage Status](https://coveralls.io/repos/github/rafaeljusto/gocnab/badge.svg?branch=master)](https://coveralls.io/github/rafaeljusto/gocnab?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/rafaeljusto/gocnab)](https://goreportcard.com/report/github.com/rafaeljusto/gocnab)
[![codebeat badge](https://codebeat.co/badges/b3a4c784-49db-4e3f-81f7-c35f4e35f70a)](https://codebeat.co/projects/github-com-rafaeljusto-gocnab-master)

![toglacier](https://raw.githubusercontent.com/rafaeljusto/gocnab/master/gocnab.png)

# gocnab

gocnab implements encoding and decoding of CNAB (Centro Nacional de Automação
Bancária) data as defined by [FEBRABAN](https://www.febraban.org.br/).

When marshaling it is possible to inform a struct, that will generate 1 CNAB
line, or a slice of struct to generate multiple CNAB lines. On unmarshal a
pointer to a struct, a pointer to a slice of struct or a mapper
(`map[string]interface{}` for a full CNAB file) should be used.

The library use struct tags to define the position of the field in the CNAB
content `[begin,end)`. It supports the basic attribute types `string` (uppercase
and left align), `bool` (represented by `1` or `0`), `int`, `int8`, `int16`,
`int32`, `int64`, `uint`, `uint8`, `uint16`, `uint23`, `uint64`, `float32` and
`float64` (decimal separator removed). And for custom types it is possible to
implement `gocnab.Marshaler`, `gocnab.Unmarshaler`, `encoding.TextMarshaler` and
`encoding.TextUnmarshaler` to make full use of this library.

## Install

```
go get -u github.com/rafaeljusto/gocnab
```

## Usage

For working with only a single line of the CNAB file:

```go
package main

import "github.com/rafaeljusto/gocnab"

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
		FieldB: "THIS IS A TEST",
		FieldC: 50.30,
		FieldD: 445,
		FieldE: true,
	}

	data, err := gocnab.Marshal400(e1)
	if err != nil {
		println(err)
		return
	}

	var e2 example
	if err = gocnab.Unmarshal(data, &e2); err != nil {
		println(err)
		return
	}

	println(e1 == e2)
}
```

And for the whole CNAB file:

```go
package main

import "github.com/rafaeljusto/gocnab"

type header struct {
	Identifier string `cnab:"0,1"`
	HeaderA    int    `cnab:"1,5"`
}

type content struct {
	Identifier string  `cnab:"0,1"`
	FieldA     int     `cnab:"1,20"`
	FieldB     string  `cnab:"20,50"`
	FieldC     float64 `cnab:"50,60"`
	FieldD     uint    `cnab:"60,70"`
	FieldE     bool    `cnab:"70,71"`
}

type footer struct {
	Identifier string `cnab:"0,1"`
	FooterA    string `cnab:"5,30"`
}

func main() {
	h1 := header{
		Identifier: "0",
		HeaderA:    2,
	}

	c1 := []content{
		{
			Identifier: "1",
			FieldA:     123,
			FieldB:     "THIS IS A TEXT",
			FieldC:     50.30,
			FieldD:     445,
			FieldE:     true,
		},
		{
			Identifier: "1",
			FieldA:     321,
			FieldB:     "THIS IS ANOTHER TEXT",
			FieldC:     30.50,
			FieldD:     544,
			FieldE:     false,
		},
	}

	f1 := footer{
		Identifier: "2",
		FooterA:    "FINAL TEXT",
	}

	data, err := gocnab.Marshal400(h1, c1, f1)
	if err != nil {
		println(err)
		return
	}

	var h2 header
	var c2 []content
	var f2 footer

	if err = gocnab.Unmarshal(data, map[string]interface{}{
		"0": &h2,
		"1": &c2,
		"2": &f2,
	}); err != nil {
		println(err)
		return
	}

	println(h1 == h2)
	for i := range c1 {
		println(c1[i] == c2[i])
	}
	println(f1 == f2)
}
```