package config

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
)

func MarshalIniToFile(filePath string, data interface{}) error {
	result, err := MarshalIni(data)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filePath, result, 0644)
	if err != nil {
		return err
	}

	return nil
}

func UnmarshalIniFormFile(filePath string, config interface{}) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	return UnmarshalIni(data, config)
}

func MarshalIni(config interface{}) (result []byte, err error) {
	// 解析 config 必须是一个结构类型
	structInfo := reflect.TypeOf(config)
	structInfoKind := structInfo.Kind()
	if structInfoKind != reflect.Struct {
		err = fmt.Errorf("config type must struct, %v give", structInfoKind)
		return
	}

	var data []string
	valueInfo := reflect.ValueOf(config)
	for i := 0; i < structInfo.NumField(); i++ {
		sectionField := structInfo.Field(i)
		sectionFieldType := sectionField.Type
		if sectionFieldType.Kind() != reflect.Struct {
			continue
		}
		sectionFieldVal := valueInfo.Field(i)
		// 节点名称
		tagVal := sectionField.Tag.Get("ini")
		if len(tagVal) == 0 {
			tagVal = sectionField.Name // 没有设置tag 默认使用字段名称
		}
		sectionNameStr := fmt.Sprintf("\n[%s]\n", tagVal)
		data = append(data, sectionNameStr)
		// 节点属性
		for j := 0; j < sectionFieldType.NumField(); j++ {
			keyField := sectionFieldType.Field(j)
			keyFieldTagVal := keyField.Tag.Get("ini")
			if len(keyFieldTagVal) == 0 {
				keyFieldTagVal = keyField.Name // 没有设置tag 默认使用字段名称
			}
			valField := sectionFieldVal.Field(j)
			item := fmt.Sprintf("%s=%v\n", keyFieldTagVal, valField.Interface())
			data = append(data, item)
		}
	}

	for index, v := range data {
		if index == 0 {
			// 首行去掉左边 \n
			v = strings.TrimLeft(v, "\n")
		}
		byteVal := []byte(v)
		result = append(result, byteVal...)
	}

	return
}

func UnmarshalIni(data []byte, configPtr interface{}) (err error) {
	// 解析 Interface 必须是一个指针
	typeInfo := reflect.TypeOf(configPtr)
	typeInfoKind := typeInfo.Kind()
	if typeInfoKind != reflect.Ptr {
		return fmt.Errorf("configPtr type must Ptr, %v give", typeInfoKind)
	}
	// 解析 configPtr 必须是一个结构类型的指针
	structInfo := typeInfo.Elem()
	structInfoKind := structInfo.Kind()
	if structInfoKind != reflect.Struct {
		return fmt.Errorf("configPtr type must struct, %v give", structInfoKind)
	}

	// 解析传入的数据
	iniArr := strings.Split(string(data), "\n")
	var lastFieldName string // 每次循环后的 sectionName Config中的属性
	for index, value := range iniArr {
		line := strings.TrimSpace(value)
		lineLen := len(line)
		if lineLen == 0 {
			// 如果是空行直接忽略
			continue
		}
		// 第1个字符
		firstChar := line[0]
		if firstChar == '#' || firstChar == ';' {
			// 如果是注释，直接忽略
			continue
		}
		lineNo := index + 1
		// 解析节点
		if firstChar == '[' {
			lastFieldName, err = parseSection(line, structInfo)
			if err != nil {
				err = fmt.Errorf("%v, lineNo: %d", err, lineNo)
				return
			}
			continue
		}
		// 能够匹配到节点名称，才去匹配节点下的属性
		if len(lastFieldName) > 0 {
			// 解析每个节点下的属性
			err = parseSectionItem(lastFieldName, line, configPtr)
			if err != nil {
				err = fmt.Errorf("%v, lineNo: %d", err, lineNo)
				return
			}
		}
	}
	return
}

func parseSection(line string, structInfo reflect.Type) (fieldName string, err error) {
	lineLen := len(line)
	if lineLen <= 2 {
		// 第一个字符是 [ 时，不能长度 <= 2 因为 [] 已经是2
		err = fmt.Errorf("syntas error, invalid section: %s, line len must egt 2, %d give", line, lineLen)
		return
	}
	endChar := line[lineLen-1]
	if endChar != ']' {
		// 第一个字符是 [ 时，不能长度 <= 2 因为 [] 已经是2
		err = fmt.Errorf("syntax error, invalid section: %s, endChar must ], %v give", line, string(endChar))
		return
	}
	sectionName := strings.TrimSpace(line[1 : lineLen-1])
	if len(sectionName) == 0 {
		err = fmt.Errorf("syntax error, invalid section: %s, sectionName can not be empty", line)
		return
	}
	// 循环对比 配置文件中的节点值与Config中的 tag ini值
	for i := 0; i < structInfo.NumField(); i++ {
		field := structInfo.Field(i)
		tagVal := field.Tag.Get("ini")
		if tagVal == sectionName {
			fieldName = field.Name
			break
		}
	}

	return
}

func parseSectionItem(sectionName string, line string, configPtr interface{}) (err error) {
	index := strings.Index(line, "=")
	if index == -1 {
		// 找不到赋值的 =
		err = fmt.Errorf("syntax error, assignment(=) not found, line: %s", line)
		return
	}
	if index == 0 {
		// = 号不能在第一位
		err = fmt.Errorf("syntax error, assignment(=) can't be in the first place, line: %s", line)
		return
	}
	key := strings.TrimSpace(line[:index-1]) // 属性Key
	if len(key) == 0 {
		// 没有key
		err = fmt.Errorf("syntax error, key len must be egt 1, line: %s", line)
		return
	}
	// 匹配对应的节点是否一个struct
	configPtrValue := reflect.ValueOf(configPtr)
	sectionValue := configPtrValue.Elem().FieldByName(sectionName)
	sectionType := sectionValue.Type()
	if sectionType.Kind() != reflect.Struct {
		err = fmt.Errorf("field: %s must be struct", sectionName)
		return
	}
	// 节点属性key匹配对应结构体字段属性名称
	var keyFieldName string
	for i := 0; i < sectionType.NumField(); i++ {
		field := sectionType.Field(i)
		tagValue := field.Tag.Get("ini")
		if tagValue == key {
			keyFieldName = field.Name
			break
		}
	}
	if len(keyFieldName) == 0 {
		// 没有找到对应的字段名，直接忽略
		return
	}

	// 匹配struct字段类型，并进行赋值
	fieldValue := sectionValue.FieldByName(keyFieldName)
	value := strings.TrimSpace(line[index+1:]) // 属性value
	switch fieldValueKind := fieldValue.Type().Kind(); fieldValueKind {
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		newValue, newErr := strconv.ParseInt(value, 10, 64)
		if newErr != nil {
			err = fmt.Errorf("%s setting error, line: %s", value, newErr)
			return
		}
		fieldValue.SetInt(newValue)

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		newValue, newErr := strconv.ParseUint(value, 10, 64)
		if newErr != nil {
			err = fmt.Errorf("%s setting error, line: %s", value, newErr)
			return
		}
		fieldValue.SetUint(newValue)

	case reflect.Float32, reflect.Float64:
		newValue, newErr := strconv.ParseFloat(value, 64)
		if newErr != nil {
			err = fmt.Errorf("%s.%s value: %s setting error, line: %s", sectionName, keyFieldName, value, newErr)
			return
		}
		fieldValue.SetFloat(newValue)

	default:
		err = fmt.Errorf("this type is not currently supported: %v, line: %s", fieldValueKind, line)
	}

	return
}
