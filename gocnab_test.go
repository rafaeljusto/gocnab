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
		vs            []interface{}
		expected      []byte
		expectedError error
	}{
		{
			description: "it should create a CNAB240 correctly from a struct",
			vs: []interface{}{
				struct {
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
			},
			expected: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%100s",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "")),
		},
		{
			description: "it should create a CNAB240 correctly from a slice of structs",
			vs: []interface{}{
				[]struct {
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
			},
			expected: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%100s\r\n%020d%-30s%10s%010d0000000001%-30s%-30s%100s",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "",
				321, "THIS IS ANOTHER TEST", strings.Replace(fmt.Sprintf("0%010.2f", 30.50), ".", "", -1), 644, "THIS IS A CUSTOM TYPE TEST 3", "THIS IS A CUSTOM TYPE TEST 4", "")),
		},
		{
			description: "it should create a full CNAB240 correctly from multiple inputs",
			vs: []interface{}{
				struct {
					Identifier int         `cnab:"0,1"`
					FieldA     int         `cnab:"1,20"`
					FieldB     string      `cnab:"20,50"`
					FieldC     float64     `cnab:"50,60"`
					FieldD     uint        `cnab:"60,70"`
					FieldE     bool        `cnab:"70,71"`
					FieldF     bool        `cnab:"71,80"`
					FieldG     customType1 `cnab:"80,110"`
					FieldH     customType2 `cnab:"110,140"`
					FieldI     time.Time   // should ignore fields without CNAB tag
				}{
					Identifier: 0,
					FieldA:     111,
					FieldB:     "This is something",
					FieldC:     77.70,
					FieldD:     45,
					FieldE:     false,
					FieldF:     false,
					FieldG: customType1(func() ([]byte, error) {
						return []byte("Hello 1"), nil
					}),
					FieldH: customType2(func() ([]byte, error) {
						return []byte("Hello 2"), nil
					}),
				},
				[]struct {
					Identifier int         `cnab:"0,1"`
					FieldA     int         `cnab:"1,20"`
					FieldB     string      `cnab:"20,50"`
					FieldC     float64     `cnab:"50,60"`
					FieldD     uint        `cnab:"60,70"`
					FieldE     bool        `cnab:"70,71"`
					FieldF     bool        `cnab:"71,80"`
					FieldG     customType1 `cnab:"80,110"`
					FieldH     customType2 `cnab:"110,140"`
					FieldI     time.Time   // should ignore fields without CNAB tag
				}{
					{
						Identifier: 1,
						FieldA:     123,
						FieldB:     "This is a test with a long text to check if the strip is working well",
						FieldC:     50.30,
						FieldD:     445,
						FieldE:     true,
						FieldF:     false,
						FieldG: customType1(func() ([]byte, error) {
							return []byte("This is a custom type test 1"), nil
						}),
						FieldH: customType2(func() ([]byte, error) {
							return []byte("This is a custom type test 2"), nil
						}),
					},
					{
						Identifier: 1,
						FieldA:     321,
						FieldB:     "This is another test",
						FieldC:     30.50,
						FieldD:     644,
						FieldE:     false,
						FieldF:     true,
						FieldG: customType1(func() ([]byte, error) {
							return []byte("This is a custom type test 3"), nil
						}),
						FieldH: customType2(func() ([]byte, error) {
							return []byte("This is a custom type test 4"), nil
						}),
					},
				},
			},
			expected: []byte(fmt.Sprintf("0%019d%-30s%10s%010d0000000000%-30s%-30s%100s\r\n1%019d%-30s%10s%010d1000000000%-30s%-30s%100s\r\n1%019d%-30s%10s%010d0000000001%-30s%-30s%100s\x1a",
				111, "THIS IS SOMETHING", strings.Replace(fmt.Sprintf("0%010.2f", 77.70), ".", "", -1), 45, "HELLO 1", "HELLO 2", "",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "",
				321, "THIS IS ANOTHER TEST", strings.Replace(fmt.Sprintf("0%010.2f", 30.50), ".", "", -1), 644, "THIS IS A CUSTOM TYPE TEST 3", "THIS IS A CUSTOM TYPE TEST 4", "")),
		},
		{
			description: "it should detect an invalid field format",
			vs: []interface{}{
				struct {
					FieldA int `cnab:"xxxxxxxx"`
				}{},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagFormat,
			},
		},
		{
			description: "it should detect an invalid begin range",
			vs: []interface{}{
				[]struct {
					FieldA int `cnab:"X,20"`
				}{
					{},
				},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagBeginRange,
			},
		},
		{
			description: "it should detect an invalid end range",
			vs: []interface{}{
				struct {
					FieldA int `cnab:"0,X"`
				}{},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagEndRange,
			},
		},
		{
			description: "it should detect an invalid range (negative begin)",
			vs: []interface{}{
				struct {
					FieldA int `cnab:"-1,20"`
				}{},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagRange,
			},
		},
		{
			description: "it should detect an invalid range (end before begin)",
			vs: []interface{}{
				struct {
					FieldA int `cnab:"20,0"`
				}{},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagRange,
			},
		},
		{
			description: "it should detect an invalid range (end after CNAB limit)",
			vs: []interface{}{
				struct {
					FieldA int `cnab:"0,241"`
				}{},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagRange,
			},
		},
		{
			description: "it should detect an error in MarshalCNAB",
			vs: []interface{}{
				struct {
					FieldG customType1 `cnab:"80,110"`
				}{
					FieldG: customType1(func() ([]byte, error) {
						return nil, errors.New("generic problem")
					}),
				},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldG",
				Err:   errors.New("generic problem"),
			},
		},
		{
			description: "it should detect an error in encoding.MarshalText",
			vs: []interface{}{
				struct {
					FieldH customType2 `cnab:"110,140"`
				}{
					FieldH: customType2(func() ([]byte, error) {
						return nil, errors.New("generic problem")
					}),
				},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldH",
				Err:   errors.New("generic problem"),
			},
		},
		{
			description: "it should detect an unsupported field",
			vs: []interface{}{
				struct {
					FieldJ struct{} `cnab:"140,150"`
				}{},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldJ",
				Err:   gocnab.ErrUnsupportedType,
			},
		},
		{
			description:   "it should detect an unsupported root type",
			vs:            []interface{}{10},
			expectedError: gocnab.ErrUnsupportedType,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			data, err := gocnab.Marshal240(scenario.vs...)

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
		vs            []interface{}
		expected      []byte
		expectedError error
	}{
		{
			description: "it should create a CNAB400 correctly from a struct",
			vs: []interface{}{
				struct {
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
			},
			expected: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%260s",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "")),
		},
		{
			description: "it should create a CNAB400 correctly from a slice of structs",
			vs: []interface{}{
				[]struct {
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
			},
			expected: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%260s\r\n%020d%-30s%10s%010d0000000001%-30s%-30s%260s",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "",
				321, "THIS IS ANOTHER TEST", strings.Replace(fmt.Sprintf("0%010.2f", 30.50), ".", "", -1), 644, "THIS IS A CUSTOM TYPE TEST 3", "THIS IS A CUSTOM TYPE TEST 4", "")),
		},
		{
			description: "it should create a full CNAB400 correctly from multiple inputs",
			vs: []interface{}{
				struct {
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
					FieldA: 111,
					FieldB: "This is something",
					FieldC: 77.70,
					FieldD: 45,
					FieldE: false,
					FieldF: false,
					FieldG: customType1(func() ([]byte, error) {
						return []byte("Hello 1"), nil
					}),
					FieldH: customType2(func() ([]byte, error) {
						return []byte("Hello 2"), nil
					}),
				},
				[]struct {
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
			},
			expected: []byte(fmt.Sprintf("%020d%-30s%10s%010d0000000000%-30s%-30s%260s\r\n%020d%-30s%10s%010d1000000000%-30s%-30s%260s\r\n%020d%-30s%10s%010d0000000001%-30s%-30s%260s\x1a",
				111, "THIS IS SOMETHING", strings.Replace(fmt.Sprintf("0%010.2f", 77.70), ".", "", -1), 45, "HELLO 1", "HELLO 2", "",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "",
				321, "THIS IS ANOTHER TEST", strings.Replace(fmt.Sprintf("0%010.2f", 30.50), ".", "", -1), 644, "THIS IS A CUSTOM TYPE TEST 3", "THIS IS A CUSTOM TYPE TEST 4", "")),
		},
		{
			description: "it should detect an invalid field format",
			vs: []interface{}{
				struct {
					FieldA int `cnab:"xxxxxxxx"`
				}{},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagFormat,
			},
		},
		{
			description: "it should detect an invalid begin range",
			vs: []interface{}{
				[]struct {
					FieldA int `cnab:"X,20"`
				}{
					{},
				},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagBeginRange,
			},
		},
		{
			description: "it should detect an invalid end range",
			vs: []interface{}{
				struct {
					FieldA int `cnab:"0,X"`
				}{},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagEndRange,
			},
		},
		{
			description: "it should detect an invalid range (negative begin)",
			vs: []interface{}{
				struct {
					FieldA int `cnab:"-1,20"`
				}{},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagRange,
			},
		},
		{
			description: "it should detect an invalid range (end before begin)",
			vs: []interface{}{
				struct {
					FieldA int `cnab:"20,0"`
				}{},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagRange,
			},
		},
		{
			description: "it should detect an invalid range (end after CNAB limit)",
			vs: []interface{}{
				struct {
					FieldA int `cnab:"0,401"`
				}{},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldA",
				Err:   gocnab.ErrInvalidFieldTagRange,
			},
		},
		{
			description: "it should detect an error in MarshalCNAB",
			vs: []interface{}{
				struct {
					FieldG customType1 `cnab:"80,110"`
				}{
					FieldG: customType1(func() ([]byte, error) {
						return nil, errors.New("generic problem")
					}),
				},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldG",
				Err:   errors.New("generic problem"),
			},
		},
		{
			description: "it should detect an error in encoding.MarshalText",
			vs: []interface{}{
				struct {
					FieldH customType2 `cnab:"110,140"`
				}{
					FieldH: customType2(func() ([]byte, error) {
						return nil, errors.New("generic problem")
					}),
				},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldH",
				Err:   errors.New("generic problem"),
			},
		},
		{
			description: "it should detect an unsupported field",
			vs: []interface{}{
				struct {
					FieldJ struct{} `cnab:"140,150"`
				}{},
			},
			expectedError: gocnab.FieldError{
				Field: "FieldJ",
				Err:   gocnab.ErrUnsupportedType,
			},
		},
		{
			description:   "it should detect an unsupported root type",
			vs:            []interface{}{10},
			expectedError: gocnab.ErrUnsupportedType,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			data, err := gocnab.Marshal400(scenario.vs...)

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
			data: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%260s\r\n%020d%-30s%10s%010d0000000001%-30s%-30s%260s\r\n",
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
			description: "it should unmarshal to a mapper correctly",
			data: []byte(fmt.Sprintf("0%019d%-30s%10s%010d0000000000%-30s%-30s%100s\r\n1%019d%-30s%10s%010d1000000000%-30s%-30s%100s\r\n1%019d%-30s%10s%010d0000000001%-30s%-30s%100s\x1a",
				111, "THIS IS SOMETHING", strings.Replace(fmt.Sprintf("0%010.2f", 77.70), ".", "", -1), 45, "HELLO 1", "HELLO 2", "",
				123, "THIS IS A TEST WITH A LONG TEX", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "THIS IS A CUSTOM TYPE TEST 1", "THIS IS A CUSTOM TYPE TEST 2", "",
				321, "THIS IS ANOTHER TEST", strings.Replace(fmt.Sprintf("0%010.2f", 30.50), ".", "", -1), 644, "THIS IS A CUSTOM TYPE TEST 3", "THIS IS A CUSTOM TYPE TEST 4", "")),
			v: map[string]interface{}{
				"0": &struct {
					Identifier int         `cnab:"0,1"`
					FieldA     int         `cnab:"1,20"`
					FieldB     string      `cnab:"20,50"`
					FieldC     float64     `cnab:"50,60"`
					FieldD     uint        `cnab:"60,70"`
					FieldE     bool        `cnab:"70,71"`
					FieldF     bool        `cnab:"71,80"`
					FieldG     customType3 `cnab:"80,110"`
					FieldH     customType4 `cnab:"110,140"`
					FieldI     time.Time   // should ignore fields without CNAB tag
				}{},
				"1": &[]struct {
					Identifier int         `cnab:"0,1"`
					FieldA     int         `cnab:"1,20"`
					FieldB     string      `cnab:"20,50"`
					FieldC     float64     `cnab:"50,60"`
					FieldD     uint        `cnab:"60,70"`
					FieldE     bool        `cnab:"70,71"`
					FieldF     bool        `cnab:"71,80"`
					FieldG     customType3 `cnab:"80,110"`
					FieldH     customType4 `cnab:"110,140"`
					FieldI     time.Time   // should ignore fields without CNAB tag
				}{},
			},
			expected: map[string]interface{}{
				"0": &struct {
					Identifier int         `cnab:"0,1"`
					FieldA     int         `cnab:"1,20"`
					FieldB     string      `cnab:"20,50"`
					FieldC     float64     `cnab:"50,60"`
					FieldD     uint        `cnab:"60,70"`
					FieldE     bool        `cnab:"70,71"`
					FieldF     bool        `cnab:"71,80"`
					FieldG     customType3 `cnab:"80,110"`
					FieldH     customType4 `cnab:"110,140"`
					FieldI     time.Time   // should ignore fields without CNAB tag
				}{
					Identifier: 0,
					FieldA:     111,
					FieldB:     "THIS IS SOMETHING",
					FieldC:     77.70,
					FieldD:     45,
					FieldE:     false,
					FieldF:     false,
					FieldG: customType3{
						data: "HELLO 1",
					},
					FieldH: customType4{
						data: "HELLO 2",
					},
				},
				"1": &[]struct {
					Identifier int         `cnab:"0,1"`
					FieldA     int         `cnab:"1,20"`
					FieldB     string      `cnab:"20,50"`
					FieldC     float64     `cnab:"50,60"`
					FieldD     uint        `cnab:"60,70"`
					FieldE     bool        `cnab:"70,71"`
					FieldF     bool        `cnab:"71,80"`
					FieldG     customType3 `cnab:"80,110"`
					FieldH     customType4 `cnab:"110,140"`
					FieldI     time.Time   // should ignore fields without CNAB tag
				}{
					{
						Identifier: 1,
						FieldA:     123,
						FieldB:     "THIS IS A TEST WITH A LONG TEX",
						FieldC:     50.30,
						FieldD:     445,
						FieldE:     true,
						FieldF:     false,
						FieldG: customType3{
							data: "THIS IS A CUSTOM TYPE TEST 1",
						},
						FieldH: customType4{
							data: "THIS IS A CUSTOM TYPE TEST 2",
						},
					},
					{
						Identifier: 1,
						FieldA:     321,
						FieldB:     "THIS IS ANOTHER TEST",
						FieldC:     30.50,
						FieldD:     644,
						FieldE:     false,
						FieldF:     true,
						FieldG: customType3{
							data: "THIS IS A CUSTOM TYPE TEST 3",
						},
						FieldH: customType4{
							data: "THIS IS A CUSTOM TYPE TEST 4",
						},
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
			data: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%260s\r\n%020d%-30s%10s%010d0000000001%-30s%-30s%260s\r\n",
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
			data: []byte(fmt.Sprintf("%020d%-30s%10s%010d1000000000%-30s%-30s%260s\r\n%020d%-30s%10s%010d0000000001%-30s%-30s%260s\r\n",
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
	t.Parallel()

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

func ExampleMarshal240_fullFile() {
	header := struct {
		HeaderA int `cnab:"0,5"`
	}{
		HeaderA: 2,
	}

	content := []struct {
		FieldA int     `cnab:"0,20"`
		FieldB string  `cnab:"20,50"`
		FieldC float64 `cnab:"50,60"`
		FieldD uint    `cnab:"60,70"`
		FieldE bool    `cnab:"70,71"`
	}{
		{
			FieldA: 123,
			FieldB: "This is a text",
			FieldC: 50.30,
			FieldD: 445,
			FieldE: true,
		},
		{
			FieldA: 321,
			FieldB: "This is another text",
			FieldC: 30.50,
			FieldD: 544,
			FieldE: false,
		},
	}

	footer := struct {
		FooterA string `cnab:"5,30"`
	}{
		FooterA: "Final text",
	}

	data, _ := gocnab.Marshal240(header, content, footer)

	fmt.Println(string(data))
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

func ExampleMarshal400_fullFile() {
	header := struct {
		HeaderA int `cnab:"0,5"`
	}{
		HeaderA: 2,
	}

	content := []struct {
		FieldA int     `cnab:"0,20"`
		FieldB string  `cnab:"20,50"`
		FieldC float64 `cnab:"50,60"`
		FieldD uint    `cnab:"60,70"`
		FieldE bool    `cnab:"70,71"`
	}{
		{
			FieldA: 123,
			FieldB: "This is a text",
			FieldC: 50.30,
			FieldD: 445,
			FieldE: true,
		},
		{
			FieldA: 321,
			FieldB: "This is another text",
			FieldC: 30.50,
			FieldD: 544,
			FieldE: false,
		},
	}

	footer := struct {
		FooterA string `cnab:"5,30"`
	}{
		FooterA: "Final text",
	}

	data, _ := gocnab.Marshal400(header, content, footer)

	fmt.Println(string(data))
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

	fmt.Printf("%v\n", e)
	// Output: {123 THIS IS A TEXT 50.3 445 true}
}

func ExampleUnmarshal_fullFile() {
	header := struct {
		Identifier int `cnab:"0,1"`
		HeaderA    int `cnab:"1,5"`
	}{}

	content := []struct {
		Identifier int     `cnab:"0,1"`
		FieldA     int     `cnab:"1,20"`
		FieldB     string  `cnab:"20,50"`
		FieldC     float64 `cnab:"50,60"`
		FieldD     uint    `cnab:"60,70"`
		FieldE     bool    `cnab:"70,71"`
	}{}

	footer := struct {
		Identifier int    `cnab:"0,1"`
		FooterA    string `cnab:"5,30"`
	}{}

	data := []byte("00005" + gocnab.LineBreak +
		"10000000000000000123THIS IS A TEXT 1              000000503000000004451" + gocnab.LineBreak +
		"10000000000000000321THIS IS A TEXT 2              000000305000000005440" + gocnab.LineBreak +
		"2    THIS IS THE FOOTER            " + gocnab.FinalControlCharacter)

	gocnab.Unmarshal(data, map[string]interface{}{
		"0": &header,
		"1": &content,
		"2": &footer,
	})

	fmt.Printf("%v\n%v\n%v\n", header, content, footer)
	// Output: {0 5}
	// [{1 123 THIS IS A TEXT 1 50.3 445 true} {1 321 THIS IS A TEXT 2 30.5 544 false}]
	// {2 THIS IS THE FOOTER}
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
