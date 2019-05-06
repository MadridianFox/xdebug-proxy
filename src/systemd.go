package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

const SystemD = iota

const systemdTemplate = `[Unit]
Description=Dbgp proxy
After=network.target

[Service]
Type=simple
User={{.User}}
Group={{.Group}}
ExecStart={{.Binary}} proxy
Restart=always

[Install]
WantedBy=multi-user.target
`

type InstallTemplateArgs struct {
	InstallArgs
	Binary string
}

func NewTemplate(templateType int) *template.Template {
	var templateStr string
	switch templateType {
	case SystemD:
		templateStr = systemdTemplate
	}
	return template.Must(template.New("service").Parse(templateStr))
}

func ApplyAndSave(template *template.Template, args *InstallTemplateArgs) error {
	var content bytes.Buffer
	err := template.Execute(io.Writer(&content), args)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(args.OutputFile, content.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}

func SaveConfig(args *InstallArgs) error {
	binaryPath, err := filepath.Abs(os.Args[0])
	if err != nil {
		return err
	}
	templateArgs := &InstallTemplateArgs{
		*args,
		binaryPath,
	}
	serviceTpl := NewTemplate(SystemD)
	return ApplyAndSave(serviceTpl, templateArgs)
}
