#!/bin/bash

if [[ ! "$#" -eq 1 ]]; then
    printf -- "\n[E] Invalid Number of Arguments\n\n"
    exit
fi

case "$1" in
    "start")
        podman kube play ./secrets.yaml ./play.yaml &
        ;;
    "stop")
        podman kube down ./play.yaml
        ;;
    "restart")
        podman kube down ./play.yaml
        podman kube play ./secrets.yaml ./play.yaml &
        ;;
    "build")
        podman build -t yapipt-backend --file Containerfile_Yapipt ../src/ 
        podman build -t yapipt-nginx --file Containerfile_Nginx ./
        ;;
    *)
        printf -- "\n[E] Invalid Argument\n\n"
        ;;
esac
