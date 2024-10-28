#!/bin/bash
timestamps=`date +"%Y%m%d%H%M"`
tag=$timestamps
podman build -t localhost/redisecho:1.0 -f ./Dockerfile .
podman tag localhost/redisecho:1.0 eu-frankfurt-1.ocir.io/cnmegk4mhxmt/redisecho:$tag
podman login -u cnmegk4mhxmt/flexispot --password "u8jqC{kdgi02cqY{Qt>)" eu-frankfurt-1.ocir.io
podman push eu-frankfurt-1.ocir.io/cnmegk4mhxmt/redisecho:$tag
