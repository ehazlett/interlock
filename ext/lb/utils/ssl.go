package utils

import (
	ctypes "github.com/docker/docker/api/types/container"
	"github.com/ehazlett/interlock/ext"
)

const (
	DefaultSSLBackendTLSVerify = "none"
)

func SSLEnabled(config *ctypes.Config) bool {
	if _, ok := config.Labels[ext.InterlockSSLLabel]; ok {
		return true
	}

	return false
}

func SSLOnly(config *ctypes.Config) bool {
	if _, ok := config.Labels[ext.InterlockSSLOnlyLabel]; ok {
		return true
	}

	return false
}

func SSLBackend(config *ctypes.Config) bool {
	if _, ok := config.Labels[ext.InterlockSSLBackendLabel]; ok {
		return true
	}

	return false
}

func SSLCertName(config *ctypes.Config) string {
	if v, ok := config.Labels[ext.InterlockSSLCertLabel]; ok {
		return v
	}

	return ""
}

func SSLCertKey(config *ctypes.Config) string {
	if v, ok := config.Labels[ext.InterlockSSLCertKeyLabel]; ok {
		return v
	}

	return ""
}

func SSLBackendTLSVerify(config *ctypes.Config) string {
	verify := DefaultSSLBackendTLSVerify

	if v, ok := config.Labels[ext.InterlockSSLBackendTLSVerifyLabel]; ok {
		verify = v
	}

	return verify
}
