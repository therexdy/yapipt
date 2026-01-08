#!/bin/sh

if [ ! -f /etc/nginx/certs/privkey.pem ]; then
    apk add --no-cache openssl
    mkdir -p /etc/nginx/certs
    openssl req -x509 -nodes -newkey rsa:4096 -days 365 -keyout /etc/nginx/certs/privkey.pem -out /etc/nginx/certs/fullchain.pem -subj "/C=IN/ST=Karnataka/O=Venkata/OU=STEM"
fi

nginx -g 'daemon off;'
