package config

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestUnMarshalIni(t *testing.T) {
	data, err := ioutil.ReadFile("config.ini")
	if err != nil {
		t.Errorf("read file failed, err: %s\n", err)
	}
	var config Config
	err = UnmarshalIni(data, &config)
	if err != nil {
		t.Errorf("unMarshalIni failed, err: %s\n", err)
	}

	t.Logf("data: %#v\n", data)
	t.Logf("config: %#v\n", config)
}

func TestMarshalIni(t *testing.T) {
	var config Config
	result, err := MarshalIni(config)
	if err != nil {
		t.Errorf("marshalIni failed, err: %s\n", err)
	}
	resultStr := string(result)
	fmt.Println(resultStr)
}

func TestMarshalIniToFile(t *testing.T) {
	var config Config
	err := MarshalIniToFile("test.ini", config)
	if err != nil {
		t.Errorf("marshalIni failed, err: %s\n", err)
	}
	t.Log("success\n")
}

func TestUnmarshalIniFormFile(t *testing.T) {
	var config Config
	err := UnmarshalIniFormFile("config.ini", &config)
	if err != nil {
		t.Errorf("marshalIni failed, err: %s\n", err)
	}
	t.Logf("success, conf: %#v\n", config)
}
