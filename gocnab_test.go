package gocnab_test

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/rafaeljusto/gocnab"
)

func TestMarshal240(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		description   string
		v             interface{}
		expected      []byte
		expectedError error
	}{
		{
			description: "it should create a CNAB240 correctly from a struct",
			v: struct {
				FieldA int         `cnab:"0,20"`
				FieldB string      `cnab:"20,50"`
				FieldC float64     `cnab:"50,60"`
				FieldD uint        `cnab:"60,70"`
				FieldE bool        `cnab:"70,71"`
				FieldF bool        `cnab:"71,80"`
				FieldG customType1 `cnab:"80,110"`
				FieldH customType2 `cnab:"110,140"`
				FieldI time.Time   // should ignore fields without CNAB tag
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
			expected: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%100s",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "")),
		},
		{
			description: "it should create a CNAB240 correctly from a slice of structs",
			v: []struct {
				FieldA int         `cnab:"0,20"`
				FieldB string      `cnab:"20,50"`
				FieldC float64     `cnab:"50,60"`
				FieldD uint        `cnab:"60,70"`
				FieldE bool        `cnab:"70,71"`
				FieldF bool        `cnab:"71,80"`
				FieldG customType1 `cnab:"80,110"`
				FieldH customType2 `cnab:"110,140"`
				FieldI time.Time   // should ignore fields without CNAB tag
			}{
				{
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
				{
					FieldA: 321,
					FieldB: "This is another test",
					FieldC: 30.50,
					FieldD: 644,
					FieldE: false,
					FieldF: true,
					FieldG: customType1(func() ([]byte, error) {
						return []byte("This is a custom type test 3"), nil
					}),
					FieldH: customType2(func() ([]byte, error) {
						return []byte("This is a custom type test 4"), nil
					}),
				},
			},
			expected: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%100s\n\r%020d%-30s%10s%010d0000000001%-30s%-30s%100s",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "",
				321, "THIS IS ANOTHER TEST", strings.Replace(fmt.Sprintf("0%010.2f", 30.50), ".", "", -1), 644, "THIS IS A CUSTOM TYPE TEST 3", "THIS IS A CUSTOM TYPE TEST 4", "")),
		},
		{
			description: "it should detect an invalid field format",
			v: struct {
				FieldA int `cnab:"xxxxxxxx"`
			}{},
			expected: []byte(strings.Repeat(" ", 240)),
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagFormat,
			},
		},
		{
			description: "it should detect an invalid begin range",
			v: []struct {
				FieldA int `cnab:"X,20"`
			}{
				{},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagBeginRange,
			},
		},
		{
			description: "it should detect an invalid end range",
			v: struct {
				FieldA int `cnab:"0,X"`
			}{},
			expected: []byte(strings.Repeat(" ", 240)),
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagEndRange,
			},
		},
		{
			description: "it should detect an invalid range (negative begin)",
			v: struct {
				FieldA int `cnab:"-1,20"`
			}{},
			expected: []byte(strings.Repeat(" ", 240)),
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagRange,
			},
		},
		{
			description: "it should detect an invalid range (end before begin)",
			v: struct {
				FieldA int `cnab:"20,0"`
			}{},
			expected: []byte(strings.Repeat(" ", 240)),
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagRange,
			},
		},
		{
			description: "it should detect an invalid range (end after CNAB limit)",
			v: struct {
				FieldA int `cnab:"0,241"`
			}{},
			expected: []byte(strings.Repeat(" ", 240)),
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagRange,
			},
		},
		{
			description: "it should detect an error in MarshalCNAB",
			v: struct {
				FieldG customType1 `cnab:"80,110"`
			}{
				FieldG: customType1(func() ([]byte, error) {
					return nil, errors.New("generic problem")
				}),
			},
			expected: []byte(strings.Repeat(" ", 240)),
			expectedError: gocnab.FieldError{
				Field: "FieldG",
				Err:   errors.New("generic problem"),
			},
		},
		{
			description: "it should detect an error in encoding.MarshalText",
			v: struct {
				FieldH customType2 `cnab:"110,140"`
			}{
				FieldH: customType2(func() ([]byte, error) {
					return nil, errors.New("generic problem")
				}),
			},
			expected: []byte(strings.Repeat(" ", 240)),
			expectedError: gocnab.FieldError{
				Field: "FieldH",
				Err:   errors.New("generic problem"),
			},
		},
		{
			description: "it should detect an unsupported field",
			v: struct {
				FieldJ struct{} `cnab:"140,150"`
			}{},
			expected: []byte(strings.Repeat(" ", 240)),
			expectedError: gocnab.FieldError{
				Field: "FieldJ",
				Err:   gocnab.ErrUnsupportedType,
			},
		},
		{
			description:   "it should detect an unsupported root type",
			v:             10,
			expectedError: gocnab.ErrUnsupportedType,
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
	t.Parallel()

	scenarios := []struct {
		description   string
		v             interface{}
		expected      []byte
		expectedError error
	}{
		{
			description: "it should create a CNAB400 correctly from a struct",
			v: struct {
				FieldA int         `cnab:"0,20"`
				FieldB string      `cnab:"20,50"`
				FieldC float64     `cnab:"50,60"`
				FieldD uint        `cnab:"60,70"`
				FieldE bool        `cnab:"70,71"`
				FieldF bool        `cnab:"71,80"`
				FieldG customType1 `cnab:"80,110"`
				FieldH customType2 `cnab:"110,140"`
				FieldI time.Time   // should ignore fields without CNAB tag
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
			expected: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%260s",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "")),
		},
		{
			description: "it should create a CNAB400 correctly from a slice of structs",
			v: []struct {
				FieldA int         `cnab:"0,20"`
				FieldB string      `cnab:"20,50"`
				FieldC float64     `cnab:"50,60"`
				FieldD uint        `cnab:"60,70"`
				FieldE bool        `cnab:"70,71"`
				FieldF bool        `cnab:"71,80"`
				FieldG customType1 `cnab:"80,110"`
				FieldH customType2 `cnab:"110,140"`
				FieldI time.Time   // should ignore fields without CNAB tag
			}{
				{
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
				{
					FieldA: 321,
					FieldB: "This is another test",
					FieldC: 30.50,
					FieldD: 644,
					FieldE: false,
					FieldF: true,
					FieldG: customType1(func() ([]byte, error) {
						return []byte("This is a custom type test 3"), nil
					}),
					FieldH: customType2(func() ([]byte, error) {
						return []byte("This is a custom type test 4"), nil
					}),
				},
			},
			expected: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%260s\n\r%020d%-30s%10s%010d0000000001%-30s%-30s%260s",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "",
				321, "THIS IS ANOTHER TEST", strings.Replace(fmt.Sprintf("0%010.2f", 30.50), ".", "", -1), 644, "THIS IS A CUSTOM TYPE TEST 3", "THIS IS A CUSTOM TYPE TEST 4", "")),
		},
		{
			description: "it should detect an invalid field format",
			v: struct {
				FieldA int `cnab:"xxxxxxxx"`
			}{},
			expected: []byte(strings.Repeat(" ", 400)),
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagFormat,
			},
		},
		{
			description: "it should detect an invalid begin range",
			v: []struct {
				FieldA int `cnab:"X,20"`
			}{
				{},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagBeginRange,
			},
		},
		{
			description: "it should detect an invalid end range",
			v: struct {
				FieldA int `cnab:"0,X"`
			}{},
			expected: []byte(strings.Repeat(" ", 400)),
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagEndRange,
			},
		},
		{
			description: "it should detect an invalid range (negative begin)",
			v: struct {
				FieldA int `cnab:"-1,20"`
			}{},
			expected: []byte(strings.Repeat(" ", 400)),
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagRange,
			},
		},
		{
			description: "it should detect an invalid range (end before begin)",
			v: struct {
				FieldA int `cnab:"20,0"`
			}{},
			expected: []byte(strings.Repeat(" ", 400)),
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagRange,
			},
		},
		{
			description: "it should detect an invalid range (end after CNAB limit)",
			v: struct {
				FieldA int `cnab:"0,401"`
			}{},
			expected: []byte(strings.Repeat(" ", 400)),
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagRange,
			},
		},
		{
			description: "it should detect an error in MarshalCNAB",
			v: struct {
				FieldG customType1 `cnab:"80,110"`
			}{
				FieldG: customType1(func() ([]byte, error) {
					return nil, errors.New("generic problem")
				}),
			},
			expected: []byte(strings.Repeat(" ", 400)),
			expectedError: gocnab.FieldError{
				Field: "FieldG",
				Err:   errors.New("generic problem"),
			},
		},
		{
			description: "it should detect an error in encoding.MarshalText",
			v: struct {
				FieldH customType2 `cnab:"110,140"`
			}{
				FieldH: customType2(func() ([]byte, error) {
					return nil, errors.New("generic problem")
				}),
			},
			expected: []byte(strings.Repeat(" ", 400)),
			expectedError: gocnab.FieldError{
				Field: "FieldH",
				Err:   errors.New("generic problem"),
			},
		},
		{
			description: "it should detect an unsupported field",
			v: struct {
				FieldJ struct{} `cnab:"140,150"`
			}{},
			expected: []byte(strings.Repeat(" ", 400)),
			expectedError: gocnab.FieldError{
				Field: "FieldJ",
				Err:   gocnab.ErrUnsupportedType,
			},
		},
		{
			description:   "it should detect an unsupported root type",
			v:             10,
			expectedError: gocnab.ErrUnsupportedType,
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

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		description   string
		data          []byte
		v             interface{}
		expected      interface{}
		expectedError error
	}{
		{
			description: "it should unmarshal to a struct correctly",
			data: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%100s",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "")),
			v: &struct {
				FieldA int         `cnab:"0,20"`
				FieldB string      `cnab:"20,50"`
				FieldC float64     `cnab:"50,60"`
				FieldD uint        `cnab:"60,70"`
				FieldE bool        `cnab:"70,71"`
				FieldF bool        `cnab:"71,80"`
				FieldG customType3 `cnab:"80,110"`
				FieldH customType4 `cnab:"110,140"`
				FieldI time.Time   // should ignore fields without CNAB tag
				fieldJ string      `cnab:"140,150"` // should ignore not exported fields
			}{},
			expected: &struct {
				FieldA int         `cnab:"0,20"`
				FieldB string      `cnab:"20,50"`
				FieldC float64     `cnab:"50,60"`
				FieldD uint        `cnab:"60,70"`
				FieldE bool        `cnab:"70,71"`
				FieldF bool        `cnab:"71,80"`
				FieldG customType3 `cnab:"80,110"`
				FieldH customType4 `cnab:"110,140"`
				FieldI time.Time   // should ignore fields without CNAB tag
				fieldJ string      `cnab:"140,150"` // should ignore not exported fields
			}{
				FieldA: 123,
				FieldB: "THIS IS A TEST WITH A LONG TEX",
				FieldC: 50.30,
				FieldD: 445,
				FieldE: true,
				FieldF: false,
				FieldG: customType3{
					data: "THIS IS A CUSTOM TYPE TEST 1",
				},
				FieldH: customType4{
					data: "THIS IS A CUSTOM TYPE TEST 2",
				},
			},
		},
		{
			description: "it should unmarshal to a slice of structs correctly",
			data: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%260s\n\r%020d%-30s%10s%010d0000000001%-30s%-30s%260s\n\r",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "",
				321, "THIS IS ANOTHER TEST", strings.Replace(fmt.Sprintf("0%010.2f", 30.50), ".", "", -1), 644, "THIS IS A CUSTOM TYPE TEST 3", "THIS IS A CUSTOM TYPE TEST 4", "")),
			v: &[]struct {
				FieldA int         `cnab:"0,20"`
				FieldB string      `cnab:"20,50"`
				FieldC float64     `cnab:"50,60"`
				FieldD uint        `cnab:"60,70"`
				FieldE bool        `cnab:"70,71"`
				FieldF bool        `cnab:"71,80"`
				FieldG customType3 `cnab:"80,110"`
				FieldH customType4 `cnab:"110,140"`
				FieldI time.Time   // should ignore fields without CNAB tag
			}{},
			expected: &[]struct {
				FieldA int         `cnab:"0,20"`
				FieldB string      `cnab:"20,50"`
				FieldC float64     `cnab:"50,60"`
				FieldD uint        `cnab:"60,70"`
				FieldE bool        `cnab:"70,71"`
				FieldF bool        `cnab:"71,80"`
				FieldG customType3 `cnab:"80,110"`
				FieldH customType4 `cnab:"110,140"`
				FieldI time.Time   // should ignore fields without CNAB tag
			}{
				{
					FieldA: 123,
					FieldB: "THIS IS A TEST WITH A LONG TEX",
					FieldC: 50.30,
					FieldD: 445,
					FieldE: true,
					FieldF: false,
					FieldG: customType3{
						data: "THIS IS A CUSTOM TYPE TEST 1",
					},
					FieldH: customType4{
						data: "THIS IS A CUSTOM TYPE TEST 2",
					},
				},
				{
					FieldA: 321,
					FieldB: "THIS IS ANOTHER TEST",
					FieldC: 30.50,
					FieldD: 644,
					FieldE: false,
					FieldF: true,
					FieldG: customType3{
						data: "THIS IS A CUSTOM TYPE TEST 3",
					},
					FieldH: customType4{
						data: "THIS IS A CUSTOM TYPE TEST 4",
					},
				},
			},
		},
		{
			description: "it should detect when output type is not a pointer",
			data: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%100s",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "")),
			v: struct {
				FieldA int `cnab:"0,20"`
			}{},
			expected: struct {
				FieldA int `cnab:"0,20"`
			}{},
			expectedError: gocnab.ErrUnsupportedType,
		},
		{
			description: "it should detect when output type is not a slice of struct",
			data: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%260s\n\r%020d%-30s%10s%010d0000000001%-30s%-30s%260s\n\r",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "",
				321, "THIS IS ANOTHER TEST", strings.Replace(fmt.Sprintf("0%010.2f", 30.50), ".", "", -1), 644, "THIS IS A CUSTOM TYPE TEST 3", "THIS IS A CUSTOM TYPE TEST 4", "")),
			v:             &[]int{},
			expected:      &[]int{},
			expectedError: gocnab.ErrUnsupportedType,
		},
		{
			description: "it should detect when output type is not supported",
			data: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%100s",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "")),
			v:             func() *int { var x int; return &x }(),
			expected:      func() *int { var x int; return &x }(),
			expectedError: gocnab.ErrUnsupportedType,
		},
		{
			description: "it should detect an invalid field format",
			data: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%100s",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "")),
			v: &struct {
				FieldA int `cnab:"xxxxxxxx"`
			}{},
			expected: &struct {
				FieldA int `cnab:"xxxxxxxx"`
			}{},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagFormat,
			},
		},
		{
			description: "it should detect an invalid begin range",
			data: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%260s\n\r%020d%-30s%10s%010d0000000001%-30s%-30s%260s\n\r",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "",
				321, "THIS IS ANOTHER TEST", strings.Replace(fmt.Sprintf("0%010.2f", 30.50), ".", "", -1), 644, "THIS IS A CUSTOM TYPE TEST 3", "THIS IS A CUSTOM TYPE TEST 4", "")),
			v: &[]struct {
				FieldA int `cnab:"X,20"`
			}{
				{},
			},
			expected: &[]struct {
				FieldA int `cnab:"X,20"`
			}{
				{},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagBeginRange,
			},
		},
		{
			description: "it should detect an invalid end range",
			data: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%100s",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "")),
			v: &struct {
				FieldA int `cnab:"0,X"`
			}{},
			expected: &struct {
				FieldA int `cnab:"0,X"`
			}{},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagEndRange,
			},
		},
		{
			description: "it should detect an invalid range (negative begin)",
			data: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%100s",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "")),
			v: &struct {
				FieldA int `cnab:"-1,20"`
			}{},
			expected: &struct {
				FieldA int `cnab:"-1,20"`
			}{},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagRange,
			},
		},
		{
			description: "it should detect an invalid range (end before begin)",
			data: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%100s",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "")),
			v: &struct {
				FieldA int `cnab:"20,0"`
			}{},
			expected: &struct {
				FieldA int `cnab:"20,0"`
			}{},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagRange,
			},
		},
		{
			description: "it should detect an invalid range (end after CNAB limit)",
			data: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%100s",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "")),
			v: &struct {
				FieldA int `cnab:"0,241"`
			}{},
			expected: &struct {
				FieldA int `cnab:"0,241"`
			}{},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagRange,
			},
		},
		{
			description: "it should detect an invalid value to fill a boolean field",
			data:        []byte(fmt.Sprintf("%1s%399s", "X", "")),
			v: &struct {
				FieldA bool `cnab:"0,1"`
			}{},
			expected: &struct {
				FieldA bool `cnab:"0,1"`
			}{},
			expectedError: gocnab.UnmarshalFieldError{
				Field: "FieldA",
				Data:  []byte("X"),
				Err: &strconv.NumError{
					Func: "ParseInt",
					Num:  "X",
					Err:  strconv.ErrSyntax,
				},
			},
		},
		{
			description: "it should detect an invalid value to fill a int field",
			data:        []byte(fmt.Sprintf("%1s%399s", "X", "")),
			v: &struct {
				FieldA int `cnab:"0,1"`
			}{},
			expected: &struct {
				FieldA int `cnab:"0,1"`
			}{},
			expectedError: gocnab.UnmarshalFieldError{
				Field: "FieldA",
				Data:  []byte("X"),
				Err: &strconv.NumError{
					Func: "ParseInt",
					Num:  "X",
					Err:  strconv.ErrSyntax,
				},
			},
		},
		{
			description: "it should detect an invalid value to fill a uint field",
			data:        []byte(fmt.Sprintf("%1s%399s", "X", "")),
			v: &struct {
				FieldA uint `cnab:"0,1"`
			}{},
			expected: &struct {
				FieldA uint `cnab:"0,1"`
			}{},
			expectedError: gocnab.UnmarshalFieldError{
				Field: "FieldA",
				Data:  []byte("X"),
				Err: &strconv.NumError{
					Func: "ParseUint",
					Num:  "X",
					Err:  strconv.ErrSyntax,
				},
			},
		},
		{
			description: "it should detect an invalid value to fill a float field",
			data:        []byte(fmt.Sprintf("%2s%398s", "XX", "")),
			v: &struct {
				FieldA float64 `cnab:"0,2"`
			}{},
			expected: &struct {
				FieldA float64 `cnab:"0,2"`
			}{},
			expectedError: gocnab.UnmarshalFieldError{
				Field: "FieldA",
				Data:  []byte("XX"),
				Err: &strconv.NumError{
					Func: "ParseFloat",
					Num:  "0.XX",
					Err:  strconv.ErrSyntax,
				},
			},
		},
		{
			description: "it should detect an unknown type when filling a field",
			data:        []byte(fmt.Sprintf("%1s%399s", "X", "")),
			v: &struct {
				FieldA struct{} `cnab:"0,1"`
			}{},
			expected: &struct {
				FieldA struct{} `cnab:"0,1"`
			}{},
			expectedError: gocnab.UnmarshalFieldError{
				Field: "FieldA",
				Data:  []byte("X"),
				Err:   gocnab.ErrUnsupportedType,
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			err := gocnab.Unmarshal(scenario.data, scenario.v)

			if !reflect.DeepEqual(scenario.expected, scenario.v) {
				t.Errorf("expected data “%#v” and got “%#v”", scenario.expected, scenario.v)
			}

			if !reflect.DeepEqual(scenario.expectedError, err) {
				t.Errorf("expected error “%v” and got “%v”", scenario.expectedError, err)
			}
		})
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	type testType struct {
		FieldA int         `cnab:"0,20"`
		FieldB string      `cnab:"20,50"`
		FieldC float64     `cnab:"50,60"`
		FieldD uint        `cnab:"60,70"`
		FieldE bool        `cnab:"70,71"`
		FieldF bool        `cnab:"71,80"`
		FieldG customType3 `cnab:"80,110"`
		FieldH customType4 `cnab:"110,140"`
		FieldI time.Time   // should ignore fields without CNAB tag
	}

	input := []testType{
		{
			FieldA: 123,
			FieldB: "THIS IS A TEST",
			FieldC: 50.30,
			FieldD: 445,
			FieldE: true,
			FieldF: false,
			FieldG: customType3{
				data: "THIS IS A CUSTOM TYPE TEST 1",
			},
			FieldH: customType4{
				data: "THIS IS A CUSTOM TYPE TEST 2",
			},
		},
		{
			FieldA: 321,
			FieldB: "THIS IS ANOTHER TEST",
			FieldC: 30.50,
			FieldD: 644,
			FieldE: false,
			FieldF: true,
			FieldG: customType3{
				data: "THIS IS A CUSTOM TYPE TEST 3",
			},
			FieldH: customType4{
				data: "THIS IS A CUSTOM TYPE TEST 4",
			},
		},
	}

	data, err := gocnab.Marshal400(input)
	if err != nil {
		t.Fatalf("error marshalling. details: %s", err)
	}

	var output []testType
	if err = gocnab.Unmarshal(data, &output); err != nil {
		t.Fatalf("error unmarshalling. details: %s", err)
	}

	if !reflect.DeepEqual(input, output) {
		t.Errorf("expected data “%#v” and got “%#v”", input, output)
	}
}

func TestFieldError_Error(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		description string
		err         gocnab.FieldError
		expected    string
	}{
		{
			description: "it should build the error message correctly",
			err: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagRange,
			},
			expected: "gocnab: error in field FieldA. details: invalid range in cnab tag",
		},
		{
			description: "it should detect when internal error is nil",
			err: gocnab.FieldError{
				Field: "FieldA",
			},
			expected: "gocnab: error in field FieldA. details: <nil>",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			text := scenario.err.Error()

			if scenario.expected != text {
				t.Errorf("expected text “%s” and got “%s”", scenario.expected, text)
			}
		})
	}
}

func TestUnmarshalFieldError_Error(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		description string
		err         gocnab.UnmarshalFieldError
		expected    string
	}{
		{
			description: "it should build the error message correctly",
			err: gocnab.UnmarshalFieldError{
				Field: "FieldA",
				Data:  []byte("invalid input"),
				Err:   gocnab.ErrInvalidFieldTagRange,
			},
			expected: "gocnab: error unmarshaling in field FieldA with data “invalid input”. details: invalid range in cnab tag",
		},
		{
			description: "it should detect when internal data and error are nil",
			err: gocnab.UnmarshalFieldError{
				Field: "FieldA",
			},
			expected: "gocnab: error unmarshaling in field FieldA with data “<nil>”. details: <nil>",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			text := scenario.err.Error()

			if scenario.expected != text {
				t.Errorf("expected text “%s” and got “%s”", scenario.expected, text)
			}
		})
	}
}

func ExampleMarshal240() {
	e := struct {
		FieldA int     `cnab:"0,20"`
		FieldB string  `cnab:"20,50"`
		FieldC float64 `cnab:"50,60"`
		FieldD uint    `cnab:"60,70"`
		FieldE bool    `cnab:"70,71"`
	}{
		FieldA: 123,
		FieldB: "This is a text",
		FieldC: 50.30,
		FieldD: 445,
		FieldE: true,
	}

	data, _ := gocnab.Marshal240(e)

	fmt.Println(string(data))
	// Output: 00000000000000000123THIS IS A TEXT                000000503000000004451
}

func ExampleMarshal400() {
	e := struct {
		FieldA int     `cnab:"0,20"`
		FieldB string  `cnab:"20,50"`
		FieldC float64 `cnab:"50,60"`
		FieldD uint    `cnab:"60,70"`
		FieldE bool    `cnab:"70,71"`
	}{
		FieldA: 123,
		FieldB: "This is a text",
		FieldC: 50.30,
		FieldD: 445,
		FieldE: true,
	}

	data, _ := gocnab.Marshal400(e)

	fmt.Println(string(data))
	// Output: 00000000000000000123THIS IS A TEXT                000000503000000004451
}

func ExampleUnmarshal() {
	var e struct {
		FieldA int     `cnab:"0,20"`
		FieldB string  `cnab:"20,50"`
		FieldC float64 `cnab:"50,60"`
		FieldD uint    `cnab:"60,70"`
		FieldE bool    `cnab:"70,71"`
	}

	data := []byte("00000000000000000123THIS IS A TEXT                000000503000000004451")
	gocnab.Unmarshal(data, &e)

	fmt.Printf("%v\n\r", e)
	// Output: {123 THIS IS A TEXT 50.3 445 true}
}

type customType1 func() ([]byte, error)

func (c customType1) MarshalCNAB() ([]byte, error) {
	return c()
}

type customType2 func() ([]byte, error)

func (c customType2) MarshalText() ([]byte, error) {
	return c()
}

type customType3 struct {
	data string
	err  error
}

func (c customType3) MarshalCNAB() ([]byte, error) {
	return []byte(c.data), c.err
}

func (c *customType3) UnmarshalCNAB(data []byte) error {
	c.data = strings.TrimSpace(string(data))
	return c.err
}

type customType4 struct {
	data string
	err  error
}

func (c *customType4) UnmarshalText(data []byte) error {
	c.data = strings.TrimSpace(string(data))
	return c.err
}

func (c customType4) MarshalCNAB() ([]byte, error) {
	return []byte(c.data), c.err
}
