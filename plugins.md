# Interlock Plugins
Plugins allow for extending Interlock.  They are simply Go packages that
implement the following interface:

```
type Plugin interface {
	Info() *PluginInfo
	HandleEvent(event *dockerclient.Event) error
}
```

To create a plugin, use the [example](https://github.com/ehazlett/interlock/tree/master/plugins/example)
plugin as a reference.  Once you have created the plugin, add the blank import
to `plugins.go` in the `interlock` package and it will be registered upon start.
You will also need to enable it when running Interlock using the `-p <name>`
flag.

If you run into issues or have questions, please open an issue or find me on
IRC (ehazlett on freenode).
