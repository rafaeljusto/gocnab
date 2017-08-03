package gocnab_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/rafaeljusto/gocnab"
)

func TestMarshal240(t *testing.T) {
	scenarios := []struct {
		description   string
		v             interface{}
		expected      []byte
		expectedError error
	}{
		{
			description: "it should create a CNAB240 correctly",
			v: struct {
				FieldA int         `cnab:"0,20"`
				FieldB string      `cnab:"20,50"`
				FieldC float64     `cnab:"50,60"`
				FieldD uint        `cnab:"60,70"`
				FieldE bool        `cnab:"70,71"`
				FieldF bool        `cnab:"71,80"`
				FieldG customType1 `cnab:"80,110"`
				FieldH customType2 `cnab:"110,140"`
			}{
				FieldA: 123,
				FieldB: "This is a test with a long text to check if the strip is working well",
				FieldC: 50.30,
				FieldD: 445,
				FieldE: true,
				FieldF: false,
				FieldG: customType1(func() ([]byte, error) {
					return []byte("This is a custom type test 1"), nil
				}),
				FieldH: customType2(func() ([]byte, error) {
					return []byte("This is a custom type test 2"), nil
				}),
			},
			expected: []byte(fmt.Sprintf("%020d%30s%10s%010d1        0%30s%30s%100s",
				123, "This is a test with a long tex", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "This is a custom type test 1", "This is a custom type test 2", "")),
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			data, err := gocnab.Marshal240(scenario.v)

			if !reflect.DeepEqual(scenario.expected, data) {
				expectedStr := "<nil>"
				if scenario.expected != nil {
					expectedStr = string(scenario.expected)
				}

				dataStr := "<nil>"
				if data != nil {
					dataStr = string(data)
				}

				t.Errorf("expected data “%s” and got “%s”", expectedStr, dataStr)
			}

			if !reflect.DeepEqual(scenario.expectedError, err) {
				t.Errorf("expected error “%v” and got “%v”", scenario.expectedError, err)
			}
		})
	}
}

func TestMarshal400(t *testing.T) {
	scenarios := []struct {
		description   string
		v             interface{}
		expected      []byte
		expectedError error
	}{
		{
			description: "it should create a CNAB400 correctly",
			v: struct {
				FieldA int         `cnab:"0,20"`
				FieldB string      `cnab:"20,50"`
				FieldC float64     `cnab:"50,60"`
				FieldD uint        `cnab:"60,70"`
				FieldE bool        `cnab:"70,71"`
				FieldF bool        `cnab:"71,80"`
				FieldG customType1 `cnab:"80,110"`
				FieldH customType2 `cnab:"110,140"`
			}{
				FieldA: 123,
				FieldB: "This is a test with a long text to check if the strip is working well",
				FieldC: 50.30,
				FieldD: 445,
				FieldE: true,
				FieldF: false,
				FieldG: customType1(func() ([]byte, error) {
					return []byte("This is a custom type test 1"), nil
				}),
				FieldH: customType2(func() ([]byte, error) {
					return []byte("This is a custom type test 2"), nil
				}),
			},
			expected: []byte(fmt.Sprintf("%020d%30s%10s%010d1        0%30s%30s%260s",
				123, "This is a test with a long tex", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "This is a custom type test 1", "This is a custom type test 2", "")),
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			data, err := gocnab.Marshal400(scenario.v)

			if !reflect.DeepEqual(scenario.expected, data) {
				expectedStr := "<nil>"
				if scenario.expected != nil {
					expectedStr = string(scenario.expected)
				}

				dataStr := "<nil>"
				if data != nil {
					dataStr = string(data)
				}

				t.Errorf("expected data “%s” and got “%s”", expectedStr, dataStr)
			}

			if !reflect.DeepEqual(scenario.expectedError, err) {
				t.Errorf("expected error “%v” and got “%v”", scenario.expectedError, err)
			}
		})
	}
}

type customType1 func() ([]byte, error)

func (c customType1) MarshalCNAB() ([]byte, error) {
	return c()
}

type customType2 func() ([]byte, error)

func (c customType2) MarshalText() ([]byte, error) {
	return c()
}
