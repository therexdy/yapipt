#!/bin/bash

if [[ ! "$#" -eq 1 ]]; then
    printf -- "\n[E] Invalid Number of Arguments\n\n"
    exit
fi

case "$1" in
    "start")
        podman kube play ./combined.yaml &
        ;;
    "stop")
        podman kube down ./combined.yaml
        ;;
    "restart")
        podman kube down ./combined.yaml
        podman kube play ./combined.yaml &
        ;;
    "build")
        podman kube down ./combined.yaml
        podman rmi yapipt-backend:latest
        podman rmi yapipt-nginx:latest
        podman build -t yapipt-backend --file Containerfile_Yapipt ../src/ 
        podman build -t yapipt-nginx --file Containerfile_Nginx ./
        yes | podman image prune
        ;;
    "buildrun")
        podman kube down ./combined.yaml
        podman rmi yapipt-backend:latest
        podman rmi yapipt-nginx:latest
        podman build -t yapipt-backend --file Containerfile_Yapipt ../src/ 
        podman build -t yapipt-nginx --file Containerfile_Nginx ./
        podman kube play ./combined.yaml &
        yes | podman image prune
        ;;
    *)
        printf -- "\n[E] Invalid Argument\n\n"
        ;;
esac
