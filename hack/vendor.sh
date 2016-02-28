#!/usr/bin/env bash
set -e

rm -rf vendor/
source 'hack/.vendor-helpers.sh'

clone git github.com/BurntSushi/toml master
clone git github.com/Sirupsen/logrus master
clone git github.com/codegangsta/cli master
clone git github.com/prometheus/client_golang master
clone git github.com/samalba/dockerclient master
clone git github.com/ehazlett/ttlcache master
clone git github.com/beorn7/perks master
clone git github.com/docker/docker master
clone git github.com/docker/go-units master
clone git github.com/docker/libkv master
clone git github.com/coreos/etcd master
clone git github.com/golang/protobuf master
clone git github.com/hashicorp/consul master
clone git github.com/hashicorp/serf master
clone git github.com/prometheus/client_model master
clone git github.com/prometheus/common master
clone git github.com/prometheus/procfs master
clone git github.com/golang/net master
clone git github.com/ugorji/go master
