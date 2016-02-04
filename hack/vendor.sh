#!/usr/bin/env bash
set -e

rm -rf vendor/
source 'hack/.vendor-helpers.sh'

clone git github.com/BurntSushi/toml master
clone git github.com/Sirupsen/logrus master
clone git github.com/codegangsta/cli master
clone git github.com/diegobernardes/ttlcache master
clone git github.com/prometheus/client_golang master
clone git github.com/samalba/dockerclient master
