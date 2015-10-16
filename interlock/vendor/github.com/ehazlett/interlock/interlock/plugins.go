package main

// interlock plugins
import (
	_ "github.com/ehazlett/interlock/plugins/example"
	_ "github.com/ehazlett/interlock/plugins/haproxy"
	_ "github.com/ehazlett/interlock/plugins/nginx"
	_ "github.com/ehazlett/interlock/plugins/stats"
)
