#!/bin/sh
pods=$(kubectl -n idcs-system get pods  | awk '{print $1}' ) 
for i in $pods
do
   kubectl -n idcs-system exec -it $i -c vault-agent -- sh -c "cd /vault/secrets && cat ca.pem"
done