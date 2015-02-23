package main

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/ehazlett/interlock"
)

type PluginCmd struct {
	Path string
}

func NewPluginCmd(path string) *PluginCmd {
	return &PluginCmd{
		Path: path,
	}
}

func (c *PluginCmd) Exec(input *interlock.PluginInput) (*interlock.PluginOutput, error) {
	outJson, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(c.Path)
	pipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	if _, err := pipe.Write(outJson); err != nil {
		return nil, err
	}

	var pluginOutput interlock.PluginOutput
	if err := json.NewDecoder(stdout).Decode(&pluginOutput); err != nil {
		fmt.Println(err)
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return &pluginOutput, nil
}
