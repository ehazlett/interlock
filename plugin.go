package interlock

import (
	"encoding/json"
	"os"
)

type (
	Plugin struct {
		Name    string `json:"name,omitempty"`
		Version string `json:"version,omitempty"`
		Author  string `json:"author,omitempty"`
		Url     string `json:"url,omitempty"`
		Path    string `json:"-"`
	}

	PluginInput struct {
		Command string `json:"command,omitempty"`
		Data    []byte `json:"data,omitempty"`
	}

	PluginOutput struct {
		Plugin  Plugin `json:"plugin,omitempty"`
		Command string `json:"command,omitempty"`
		Output  []byte `json:"output,omitempty"`
		Error   []byte `json:"error,omitempty"`
	}
)

// GetPluginInput deserializes a PluginInput from a JSON stream on stdin.
// It is a helper function to be used by Go plugins.
func GetPluginInput() (*PluginInput, error) {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}

	if (fi.Mode() & os.ModeCharDevice) == 0 {
		var input PluginInput
		if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
			return nil, err
		}
		return &input, nil
	} else { // if not coming from a pipe return nil, nil
		return nil, nil
	}
}

// GetPluginOutput deserializes a PluginOutput from a JSON stream on stdout.
// It is a helper function to be used by Go plugins.
func GetPluginOutput() (*PluginOutput, error) {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return nil, err
	}

	if (fi.Mode() & os.ModeCharDevice) == 0 {
		var output PluginOutput
		if err := json.NewDecoder(os.Stdout).Decode(&output); err != nil {
			return nil, err
		}
		return &output, nil
	} else { // if not coming from a pipe return nil, nil
		return nil, nil
	}
}

// SendPluginOutput sends the PluginOutput as JSON to stdout.  It is a helper
// function to be used by Go plugins.
func SendPluginOutput(out *PluginOutput) error {
	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		return err
	}
	return nil
}
