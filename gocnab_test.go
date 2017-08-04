package gocnab_test

import (
	"errors"
	"fmt"
	"reflect"
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
			expected: []byte(fmt.Sprintf("%020d%30s%10s%010d1        0%30s%30s%100s",
				123, "This is a test with a long tex", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "This is a custom type test 1", "This is a custom type test 2", "")),
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
			expected: []byte(fmt.Sprintf("%020d%30s%10s%010d1        0%30s%30s%100s\n%020d%30s%10s%010d0        1%30s%30s%100s\n",
				123, "This is a test with a long tex", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "This is a custom type test 1", "This is a custom type test 2", "",
				321, "This is another test", strings.Replace(fmt.Sprintf("0%010.2f", 30.50), ".", "", -1), 644, "This is a custom type test 3", "This is a custom type test 4", "")),
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
			expected: []byte(fmt.Sprintf("%020d%30s%10s%010d1        0%30s%30s%260s",
				123, "This is a test with a long tex", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "This is a custom type test 1", "This is a custom type test 2", "")),
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
			expected: []byte(fmt.Sprintf("%020d%30s%10s%010d1        0%30s%30s%260s\n%020d%30s%10s%010d0        1%30s%30s%260s\n",
				123, "This is a test with a long tex", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "This is a custom type test 1", "This is a custom type test 2", "",
				321, "This is another test", strings.Replace(fmt.Sprintf("0%010.2f", 30.50), ".", "", -1), 644, "This is a custom type test 3", "This is a custom type test 4", "")),
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
			data: []byte(fmt.Sprintf("%020d%30s%10s%010d1        0%30s%30s%100s",
				123, "This is a test with a long tex", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "This is a custom type test 1", "This is a custom type test 2", "")),
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
			}{
				FieldA: 123,
				FieldB: "This is a test with a long tex",
				FieldC: 50.30,
				FieldD: 445,
				FieldE: true,
				FieldF: false,
				FieldG: customType3{
					data: "This is a custom type test 1",
				},
				FieldH: customType4{
					data: "This is a custom type test 2",
				},
			},
		},
		{
			description: "it should unmarshal to a slice of structs correctly",
			data: []byte(fmt.Sprintf("%020d%30s%10s%010d1        0%30s%30s%100s\n%020d%30s%10s%010d0        1%30s%30s%100s\n",
				123, "This is a test with a long tex", strings.Replace(fmt.Sprintf("0%010.2f", 50.30), ".", "", -1), 445, "This is a custom type test 1", "This is a custom type test 2", "",
				321, "This is another test", strings.Replace(fmt.Sprintf("0%010.2f", 30.50), ".", "", -1), 644, "This is a custom type test 3", "This is a custom type test 4", "")),
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
					FieldB: "This is a test with a long tex",
					FieldC: 50.30,
					FieldD: 445,
					FieldE: true,
					FieldF: false,
					FieldG: customType3{
						data: "This is a custom type test 1",
					},
					FieldH: customType4{
						data: "This is a custom type test 2",
					},
				},
				{
					FieldA: 321,
					FieldB: "This is another test",
					FieldC: 30.50,
					FieldD: 644,
					FieldE: false,
					FieldF: true,
					FieldG: customType3{
						data: "This is a custom type test 3",
					},
					FieldH: customType4{
						data: "This is a custom type test 4",
					},
				},
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
			FieldB: "This is a test",
			FieldC: 50.30,
			FieldD: 445,
			FieldE: true,
			FieldF: false,
			FieldG: customType3{
				data: "This is a custom type test 1",
			},
			FieldH: customType4{
				data: "This is a custom type test 2",
			},
		},
		{
			FieldA: 321,
			FieldB: "This is another test",
			FieldC: 30.50,
			FieldD: 644,
			FieldE: false,
			FieldF: true,
			FieldG: customType3{
				data: "This is a custom type test 3",
			},
			FieldH: customType4{
				data: "This is a custom type test 4",
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
