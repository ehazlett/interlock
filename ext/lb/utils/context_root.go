package utils

import (
	"github.com/ehazlett/interlock/ext"
)

func ContextRoot(labels map[string]string) string {
	if v, ok := labels[ext.InterlockContextRootLabel]; ok {
		return v
	}

	return ""
}

func ContextRootRewrite(labels map[string]string) bool {
	if _, ok := labels[ext.InterlockContextRootRewriteLabel]; ok {
		return true
	}

	return false
}
