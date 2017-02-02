#!/bin/bash
TAG=1.4.0-nrcc make build-cont
TAG=1.4.0-nrcc make build-image
docker image tag ehazlett/interlock:1.4.0-nrcc interlock:1.4.0-nrcc
docker rmi ehazlett/interlock:1.4.0-nrcc

