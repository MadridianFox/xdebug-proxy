package main

import "testing"

func TestParseConfig(t *testing.T) {
	var data = `
xdebug-address: 0.0.0.0:9000
registry-address: 0.0.0.0:9001
predefined:
  - idekey: koryukov
    address: 127.0.0.1:9050
`
	config, err := parseConfig([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if config.XdebugAddr != "0.0.0.0:9000" {
		t.Error("xdebug-address is bad")
	}
	if config.RegistryAddr != "0.0.0.0:9001" {
		t.Error("registry-address is bad")
	}
	if len(config.Predefined) != 1 {
		t.Error("predefined list lenght os bad")
	}
	if config.Predefined[0].Idekey != "koryukov" {
		t.Error("idekey of first predefined item is bad")
	}
	if config.Predefined[0].Address != "127.0.0.1:9050" {
		t.Error("address of first predefined item is bad")
	}
}
