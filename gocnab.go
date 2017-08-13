// Package gocnab implements encoding and decoding of CNAB (Centro Nacional de
// Automação Bancária) as defined by FEBRABAN (Federação Brasileira de Bancos).
package gocnab

import (
	"bytes"
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// LineBreak defines the control characters at the end of each registry entry.
// It should be the hex encoded 0D0A except for the last one.
const LineBreak = "\r\n"

// FinalControlCharacter defines the control character of the last registry
// entry. It should be the hex encoded 1A.
const FinalControlCharacter = "\x1A"

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

// Marshal240 returns the CNAB 240 encoding of vs.
func Marshal240(vs ...interface{}) ([]byte, error) {
	return marshal(240, vs...)
}

// Marshal400 returns the CNAB 400 encoding of vs.
func Marshal400(vs ...interface{}) ([]byte, error) {
	return marshal(400, vs...)
}

func marshal(lineSize int, vs ...interface{}) ([]byte, error) {
	var cnab []byte

	for i, v := range vs {
		rv := reflect.ValueOf(v)

		cnabLine, err := marshalLine(lineSize, v)
		if err != nil {
			return nil, err
		}

		cnab = append(cnab, cnabLine...)

		// don't add line break symbol to the last line
		if len(vs) > 1 && i < rv.Len()-1 {
			cnab = append(cnab, []byte(LineBreak)...)
		}
	}

	if len(vs) > 1 && cnab != nil {
		cnab = append(cnab, []byte(FinalControlCharacter)...)
	}

	return cnab, nil
}

func marshalLine(lineSize int, v interface{}) ([]byte, error) {
	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.Struct:
		cnab := []byte(strings.Repeat(" ", lineSize))
		if err := marshalStruct(cnab, rv); err != nil {
			return nil, err
		}

		return cnab, nil

	case reflect.Slice:
		var cnab []byte

		for i := 0; i < rv.Len(); i++ {
			line := []byte(strings.Repeat(" ", lineSize))
			if err := marshalStruct(line, rv.Index(i)); err != nil {
				return nil, err
			}

			cnab = append(cnab, line...)

			// don't add line break symbol to the last line
			if i < rv.Len()-1 {
				cnab = append(cnab, []byte(LineBreak)...)
			}
		}

		return cnab, nil
	}

	return nil, ErrUnsupportedType
}

func marshalStruct(data []byte, v reflect.Value) error {
	structType := v.Type()
	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)
		begin, end, err := parseCNABFieldTag(structField, len(data))
		if err != nil {
			return FieldError{
				Field: structField.Name,
				Err:   err,
			}
		}

		// ignore fields without range
		if begin == 0 && end == 0 {
			continue
		}

		if err = marshalField(data, v.FieldByName(structField.Name), begin, end); err != nil {
			return FieldError{
				Field: structField.Name,
				Err:   err,
			}
		}
	}

	return nil
}

func marshalField(data []byte, v reflect.Value, begin, end int) error {
	cnabFieldSize := end - begin

	switch v.Kind() {
	case reflect.String:
		fieldContent := v.Interface().(string)
		setFieldContent(data, fieldContent, begin, end)
		return nil

	case reflect.Bool:
		fieldContent := v.Interface().(bool)
		var convertedFieldContent string
		if fieldContent {
			convertedFieldContent = "1"
		} else {
			convertedFieldContent = "0"
		}
		convertedFieldContent = fmt.Sprintf("%0"+strconv.Itoa(cnabFieldSize)+"s", convertedFieldContent)
		setFieldContent(data, convertedFieldContent, begin, end)
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fieldContent := fmt.Sprintf("%0"+strconv.Itoa(cnabFieldSize)+"d", v.Int())
		setFieldContent(data, fieldContent, begin, end)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		fieldContent := fmt.Sprintf("%0"+strconv.Itoa(cnabFieldSize)+"d", v.Uint())
		setFieldContent(data, fieldContent, begin, end)
		return nil

	case reflect.Float32, reflect.Float64:
		// replace decimal separator for nothing and add an extra 0 to fill the gap
		fieldContent := fmt.Sprintf("%0"+strconv.Itoa(cnabFieldSize)+".2f", v.Float())
		fieldContent = "0" + strings.Replace(fieldContent, ".", "", -1)
		setFieldContent(data, fieldContent, begin, end)
		return nil
	}

	marshalerType := reflect.TypeOf((*Marshaler)(nil)).Elem()
	if v.Type().Implements(marshalerType) {
		fieldContent, err := v.Interface().(Marshaler).MarshalCNAB()
		if err != nil {
			return err
		}

		setFieldContent(data, string(fieldContent), begin, end)
		return nil
	}

	textMarshalerType := reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	if v.Type().Implements(textMarshalerType) {
		fieldContent, err := v.Interface().(encoding.TextMarshaler).MarshalText()
		if err != nil {
			return err
		}

		setFieldContent(data, string(fieldContent), begin, end)
		return nil
	}

	return ErrUnsupportedType
}

func setFieldContent(data []byte, fieldContent string, begin, end int) {
	cnabFieldSize := end - begin

	// strip field if is too big for the space
	if len(fieldContent) > cnabFieldSize {
		fieldContent = fieldContent[0:cnabFieldSize]
	} else if len(fieldContent) < cnabFieldSize {
		fieldContent = fieldContent + strings.Repeat(" ", cnabFieldSize-len(fieldContent))
	}

	copy(data[begin:], strings.ToUpper(fieldContent))
}

// Unmarshal parses the CNAB-encoded data and stores the result in the value
// pointed to by v.
func Unmarshal(data []byte, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return ErrUnsupportedType
	}

	rvElem := rv.Elem()

	switch rvElem.Kind() {
	case reflect.Struct:
		return unmarshal(data, rvElem)

	case reflect.Slice:
		sliceType := rvElem.Type().Elem()
		if sliceType.Kind() != reflect.Struct {
			return ErrUnsupportedType
		}

		cnabLines := bytes.Split(data, []byte(LineBreak))
		for _, cnabLine := range cnabLines {
			if len(cnabLine) == 0 {
				continue
			}

			itemValue := reflect.New(sliceType)
			if err := unmarshal(cnabLine, itemValue.Elem()); err != nil {
				return err
			}

			rvElem.Set(reflect.Append(rvElem, itemValue.Elem()))
		}

		return nil
	}

	return ErrUnsupportedType
}

func unmarshal(data []byte, v reflect.Value) error {
	structType := v.Type()
	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)
		begin, end, err := parseCNABFieldTag(structField, len(data))
		if err != nil {
			return FieldError{
				Field: structField.Name,
				Err:   err,
			}
		}

		// ignore fields without range or not exported
		field := v.FieldByName(structField.Name)
		if (begin == 0 && end == 0) || !field.CanSet() {
			continue
		}

		if err = unmarshalField(data, field, begin, end); err != nil {
			return UnmarshalFieldError{
				Field: structField.Name,
				Data:  data[begin:end],
				Err:   err,
			}
		}
	}

	return nil
}

func unmarshalField(data []byte, v reflect.Value, begin, end int) error {
	cnabFieldStr := string(data[begin:end])
	cnabFieldStr = strings.TrimSpace(cnabFieldStr)

	switch v.Kind() {
	case reflect.String:
		v.SetString(cnabFieldStr)
		return nil

	case reflect.Bool:
		boolNumber, err := strconv.ParseInt(cnabFieldStr, 10, 64)
		if err != nil {
			return err
		}

		v.SetBool(boolNumber != 0)
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		number, err := strconv.ParseInt(cnabFieldStr, 10, 64)
		if err != nil {
			return err
		}

		v.SetInt(number)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		number, err := strconv.ParseUint(cnabFieldStr, 10, 64)
		if err != nil {
			return err
		}

		v.SetUint(number)
		return nil

	case reflect.Float32, reflect.Float64:
		numberRaw := cnabFieldStr

		// add again the dot before converting to float64
		if len(numberRaw) > 2 {
			numberRaw = numberRaw[:len(numberRaw)-2] + "." + numberRaw[len(numberRaw)-2:]
		} else {
			numberRaw = "0." + numberRaw
		}

		number, err := strconv.ParseFloat(string(numberRaw), 64)
		if err != nil {
			return err
		}

		v.SetFloat(number)
		return nil
	}

	if v.CanAddr() {
		unmarshalerType := reflect.TypeOf((*Unmarshaler)(nil)).Elem()
		if v.Addr().Type().Implements(unmarshalerType) {
			return v.Addr().Interface().(Unmarshaler).UnmarshalCNAB(data[begin:end])
		}

		textUnmarshalerType := reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
		if v.Addr().Type().Implements(textUnmarshalerType) {
			return v.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText(data[begin:end])
		}
	}

	return ErrUnsupportedType
}

func parseCNABFieldTag(structField reflect.StructField, dataSize int) (begin int, end int, err error) {
	cnabFieldOptionsRaw := structField.Tag.Get("cnab")
	if cnabFieldOptionsRaw == "" {
		return 0, 0, nil
	}

	cnabFieldOptions := strings.Split(cnabFieldOptionsRaw, ",")
	if len(cnabFieldOptions) != 2 {
		return 0, 0, ErrInvalidFieldTagFormat
	}

	begin, err = strconv.Atoi(cnabFieldOptions[0])
	if err != nil {
		return 0, 0, ErrInvalidFieldTagBeginRange
	}

	end, err = strconv.Atoi(cnabFieldOptions[1])
	if err != nil {
		return 0, 0, ErrInvalidFieldTagEndRange
	}

	if begin < 0 || end < begin || end > dataSize {
		return 0, 0, ErrInvalidFieldTagRange
	}

	return
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

// Error return a human readable representation of the field error.
func (f FieldError) Error() string {
	errStr := "<nil>"
	if f.Err != nil {
		errStr = f.Err.Error()
	}

	return fmt.Sprintf("gocnab: error in field %s. details: %s", f.Field, errStr)
}

// UnmarshalFieldError stores the error that occurred while decoding the CNAB
// data into a field.
type UnmarshalFieldError struct {
	Field string
	Data  []byte
	Err   error
}

// Error return a human readable representation of the unmarshal error.
func (u UnmarshalFieldError) Error() string {
	dataStr := "<nil>"
	if u.Data != nil {
		dataStr = string(u.Data)
	}

	errStr := "<nil>"
	if u.Err != nil {
		errStr = u.Err.Error()
	}

	return fmt.Sprintf("gocnab: error unmarshaling in field %s with data “%s”. details: %s", u.Field, dataStr, errStr)
}
