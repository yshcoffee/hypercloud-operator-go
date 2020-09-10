#/bin/sh

kubectl create secret generic registry-ca --from-file=ca.crt=./ca.crt --from-file=ca.key=./ca.key -n hypercloud4-system