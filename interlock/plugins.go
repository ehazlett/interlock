package main

// interlock plugins
import (
	_ "github.com/ehazlett/interlock/plugins/example"
	_ "github.com/ehazlett/interlock/plugins/external"
	_ "github.com/ehazlett/interlock/plugins/haproxy"
	_ "github.com/ehazlett/interlock/plugins/stats"
)
