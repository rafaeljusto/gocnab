// Package gocnab implements encoding and decoding of CNAB (Centro Nacional de
// Automação Bancária) as defined by FEBRABAN (Federação Brasileira de Bancos).
package gocnab

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	// ErrUnsupportedType raised when trying to marshal something different from a
	// struct or a slice.
	ErrUnsupportedType = errors.New("gocnab: unsupported type")

	// ErrInvalidFieldTagFormat CNAB field tag doesn't follow the expected format.
	ErrInvalidFieldTagFormat = errors.New("invalid field tag format")

	// ErrInvalidFieldTagBeginRange begin range isn't a valid number in the CNAB
	// tag.
	ErrInvalidFieldTagBeginRange = errors.New("invalid begin range in cnab tag")

	// ErrInvalidFieldTagEndRange end range isn't a valid number in the CNAB tag.
	ErrInvalidFieldTagEndRange = errors.New("invalid end range in cnab tag")

	// ErrInvalidFieldTagRange ranges don't have consistency with the desired
	// encoding in the CNAB tag.
	ErrInvalidFieldTagRange = errors.New("invalid range in cnab tag")
)

// Marshal240 returns the CNAB 240 encoding of v.
func Marshal240(v interface{}) ([]byte, error) {
	switch reflect.ValueOf(v).Kind() {
	case reflect.Struct:
		cnab240 := []byte(strings.Repeat(" ", 240))
		err := marshal(cnab240, reflect.ValueOf(v))
		return cnab240, err

	case reflect.Slice:
		var cnab240 []byte

		sliceValue := reflect.ValueOf(v)
		for i := 0; i < sliceValue.Len(); i++ {
			cnab240Line := []byte(strings.Repeat(" ", 240))
			err := marshal(cnab240Line, sliceValue.Index(i))
			if err != nil {
				return nil, err
			}

			cnab240 = append(cnab240, cnab240Line...)
			cnab240 = append(cnab240, byte('\n'))
		}

		return cnab240, nil
	}

	return nil, ErrUnsupportedType
}

// Marshal400 returns the CNAB 400 encoding of v.
func Marshal400(v interface{}) ([]byte, error) {
	switch reflect.ValueOf(v).Kind() {
	case reflect.Struct:
		cnab400 := []byte(strings.Repeat(" ", 400))
		err := marshal(cnab400, reflect.ValueOf(v))
		return cnab400, err

	case reflect.Slice:
		var cnab400 []byte

		sliceValue := reflect.ValueOf(v)
		for i := 0; i < sliceValue.Len(); i++ {
			cnab400Line := []byte(strings.Repeat(" ", 400))
			err := marshal(cnab400Line, sliceValue.Index(i))
			if err != nil {
				return nil, err
			}

			cnab400 = append(cnab400, cnab400Line...)
			cnab400 = append(cnab400, byte('\n'))
		}

		return cnab400, nil
	}

	return nil, ErrUnsupportedType
}

func marshal(cnab []byte, v reflect.Value) error {
	structType := v.Type()
	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)
		cnabFieldOptionsRaw := structField.Tag.Get("cnab")
		if cnabFieldOptionsRaw == "" {
			continue
		}

		cnabFieldOptions := strings.Split(cnabFieldOptionsRaw, ",")
		if len(cnabFieldOptions) != 2 {
			return FieldError{
				Field: structField.Name,
				Err:   ErrInvalidFieldTagFormat,
			}
		}

		begin, err := strconv.Atoi(cnabFieldOptions[0])
		if err != nil {
			return FieldError{
				Field: structField.Name,
				Err:   ErrInvalidFieldTagBeginRange,
			}
		}

		end, err := strconv.Atoi(cnabFieldOptions[1])
		if err != nil {
			return FieldError{
				Field: structField.Name,
				Err:   ErrInvalidFieldTagEndRange,
			}
		}

		if begin < 0 || end < begin || end > len(cnab) {
			return FieldError{
				Field: structField.Name,
				Err:   ErrInvalidFieldTagRange,
			}
		}

		if err = marshalField(cnab, v.FieldByName(structField.Name), begin, end); err != nil {
			return FieldError{
				Field: structField.Name,
				Err:   err,
			}
		}
	}

	return nil
}

func marshalField(cnab []byte, v reflect.Value, begin, end int) error {
	cnabFieldSize := end - begin

	switch v.Kind() {
	case reflect.String:
		fieldContent := v.Interface().(string)
		setFieldContent(cnab, fieldContent, begin, end)
		return nil

	case reflect.Bool:
		fieldContent := v.Interface().(bool)
		var convertedFieldContent string
		if fieldContent {
			convertedFieldContent = "1"
		} else {
			convertedFieldContent = "0"
		}
		setFieldContent(cnab, convertedFieldContent, begin, end)
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fieldContent := fmt.Sprintf("%0"+strconv.Itoa(cnabFieldSize)+"d", v.Int())
		setFieldContent(cnab, fieldContent, begin, end)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		fieldContent := fmt.Sprintf("%0"+strconv.Itoa(cnabFieldSize)+"d", v.Uint())
		setFieldContent(cnab, fieldContent, begin, end)
		return nil

	case reflect.Float32, reflect.Float64:
		// replace decimal separator for nothing and add an extra 0 to fill the gap
		fieldContent := fmt.Sprintf("%0"+strconv.Itoa(cnabFieldSize)+".2f", v.Float())
		fieldContent = "0" + strings.Replace(fieldContent, ".", "", -1)
		setFieldContent(cnab, fieldContent, begin, end)
		return nil
	}

	marshalerType := reflect.TypeOf((*Marshaler)(nil)).Elem()
	if v.Type().Implements(marshalerType) {
		fieldContent, err := v.Interface().(Marshaler).MarshalCNAB()
		if err != nil {
			return err
		}

		setFieldContent(cnab, string(fieldContent), begin, end)
		return nil
	}

	textMarshalerType := reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	if v.Type().Implements(textMarshalerType) {
		fieldContent, err := v.Interface().(encoding.TextMarshaler).MarshalText()
		if err != nil {
			return err
		}

		setFieldContent(cnab, string(fieldContent), begin, end)
		return nil
	}

	return ErrUnsupportedType
}

func setFieldContent(cnab []byte, fieldContent string, begin, end int) {
	cnabFieldSize := end - begin

	// strip field if is too big for the space
	if len(fieldContent) > cnabFieldSize {
		fieldContent = fieldContent[0:cnabFieldSize]
	} else if len(fieldContent) < cnabFieldSize {
		fieldContent = strings.Repeat(" ", cnabFieldSize-len(fieldContent)) + fieldContent
	}

	copy(cnab[begin:], fieldContent)
}

// Unmarshal parses the CNAB-encoded data and stores the result in the value
// pointed to by v.
func Unmarshal(data []byte, v interface{}) error {
	return nil
}

// Marshaler is the interface implemented by types that can marshal themselves
// into valid string representation.
type Marshaler interface {
	MarshalCNAB() ([]byte, error)
}

// Unmarshaler is the interface implemented by types that can unmarshal a string
// representation description of themselves. UnmarshalCNAB must copy the CNAB
// data if it wishes to retain the data after returning.
type Unmarshaler interface {
	UnmarshalCNAB([]byte) error
}

// FieldError problem detected in a field tag containing CNAB options or when
// marshalling the field itself.
type FieldError struct {
	Field string
	Err   error
}

// Error return a human readable representation of the field in tag error.
func (f FieldError) Error() string {
	errStr := "<nil>"
	if f.Err != nil {
		errStr = f.Err.Error()
	}

	return fmt.Sprintf("gocnab: error in field %s. details: %s", f.Field, errStr)
}
