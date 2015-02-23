package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"text/tabwriter"

	"github.com/codegangsta/cli"
	"github.com/ehazlett/interlock"
)

var appCommands = []cli.Command{
	{
		Name:   "ls",
		Usage:  "list available plugins",
		Action: listPlugins,
	},
}

func printErr(err error) {
	fmt.Fprintf(os.Stdout, "%s\n", err.Error())
	os.Exit(1)
}

func getPluginInfo(path string, w io.Writer, wg *sync.WaitGroup, ec chan error) {
	defer wg.Done()

	p := NewPluginCmd(path)
	input := &interlock.PluginInput{
		Command: "info",
		Args:    nil,
	}
	out, err := p.Exec(input)
	if err != nil {
		ec <- err
		return
	}

	if out != nil {
		if out.Error != nil {
			ec <- fmt.Errorf(string(out.Error))
			return
		}

		if out.Output != nil {
			var plugin interlock.Plugin
			if err := json.Unmarshal(out.Output, &plugin); err != nil {
				ec <- err
				return
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", plugin.Name,
				plugin.Version, plugin.Author, plugin.Url)
		}
	}
}

func listPlugins(c *cli.Context) {
	pluginPath := c.GlobalString("plugin-path")

	// make sure dir exists
	if err := os.MkdirAll(pluginPath, 0700); err != nil {
		printErr(err)
	}

	plugins, err := ioutil.ReadDir(pluginPath)
	if err != nil {
		printErr(err)
	}

	w := tabwriter.NewWriter(os.Stdout, 5, 1, 3, ' ', 0)

	fmt.Fprintln(w, "NAME\tVERSION\tAUTHOR\tURL")

	wg := &sync.WaitGroup{}
	ec := make(chan error)

	for _, fi := range plugins {
		wg.Add(1)
		go getPluginInfo(filepath.Join(pluginPath, fi.Name()), w, wg, ec)
	}

	wg.Wait()

	w.Flush()
}
