#! /bin/bash

openssl req -x509 -nodes -days 3650 -newkey rsa:4096 \
        -keyout ca.key -out ca.crt \
        -subj "/C=KR/ST=Seoul/L=Seoul/O=Tmax"