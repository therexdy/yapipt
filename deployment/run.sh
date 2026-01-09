#!/bin/bash

if [[ ! "$#" -eq 1 ]]; then
    printf -- "\n[E] Invalid Number of Arguments\n\n"
    exit
fi

case "$1" in
    "start")
        podman kube play ./secrets.yaml ./play.yaml &
        if exec podman kube play ./secrets.yaml ./play.yaml ; then
            printf -- "\n[I] Started\n\n"
        elif cat ./secrets.yaml ./play.yaml | podman kube play - ; then
            printf -- "\n[I] Started with fall back\n\n"
        fi
        ;;
    "stop")
        podman kube down ./play.yaml
        ;;
    "restart")
        podman kube down ./play.yaml
        if exec podman kube play ./secrets.yaml ./play.yaml ; then
            printf -- "\n[I] Started\n\n"
        elif cat ./secrets.yaml ./play.yaml | podman kube play - ; then
            printf -- "\n[I] Started with fall back\n\n"
        fi
        ;;
    "build")
        podman build -t yapipt-backend --file Containerfile_Yapipt ../src/ 
        podman build -t yapipt-nginx --file Containerfile_Nginx ./
        ;;
    *)
        printf -- "\n[E] Invalid Argument\n\n"
        ;;
esac
