package proxy

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"golang.org/x/net/html/charset"
	"regexp"
)

const CommandInit = "proxyinit"
const CommandStop = "proxystop"

type DebugPacket struct {
	Idekey string `xml:"idekey,attr"`
}

type ProxyXmlMessage struct {
	XMLName xml.Name
	Success int8   `xml:"success,attr"`
	Idekey  string `xml:"idekey,attr"`
	Error   string `xml:"error>message"`
}

type IdeCommand struct {
	Name   string
	Idekey string
	Port   string
}

func GetIdekey(packet []byte) (string, error) {
	var debugPacket DebugPacket
	decoder := xml.NewDecoder(bytes.NewReader(packet))
	decoder.CharsetReader = charset.NewReaderLabel
	err := decoder.Decode(&debugPacket)
	if err != nil {
		return "", err
	}
	return debugPacket.Idekey, nil
}

func CreateInitMessage(messageType string, idekey string, error string) ([]byte, error) {
	var success int8
	if error == "" {
		success = 1
	} else {
		success = 0
	}
	message := ProxyXmlMessage{
		XMLName: xml.Name{Local: messageType},
		Success: success,
		Idekey:  idekey,
		Error:   error,
	}
	xmlMessage, err := xml.Marshal(message)
	return xmlMessage, err
}

func extractByRegexp(pattern string, data string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	groups := re.FindStringSubmatch(data)
	if len(groups) < 2 {
		return "", errors.New(fmt.Sprintf(`no groups matched in "%s" by pattern "%s"`, data, pattern))
	}
	return groups[1], nil
}

func ParseIdeCommand(commandString string) (*IdeCommand, error) {
	ideCommand := &IdeCommand{}
	var err error
	ideCommand.Name, err = extractByRegexp(`^([^\s]+)`, commandString)
	if err != nil {
		return ideCommand, err
	}
	ideCommand.Idekey, err = extractByRegexp(`-k ([^\s]+)`, commandString)
	if err != nil {
		return ideCommand, err
	}
	if ideCommand.Name == CommandInit {
		ideCommand.Port, err = extractByRegexp(`-p ([^\s]+)`, commandString)
		if err != nil {
			return ideCommand, err
		}
	}
	return ideCommand, nil
}
