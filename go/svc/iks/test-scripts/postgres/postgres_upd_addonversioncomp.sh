#! /bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

#K8SPATCHVERSION i.e. 1.27.5
#IMINAME i.e. iks-u22-cd-cp-1.27.5-23-9-18

# DBaaS Vault config https://internal-placeholder.com/ui/vault/secrets/secret/show/dbaas/psql-iks/customer

if [ "$#" -ne 1 ]; then
    echo "Incorrect number of arguments, need to pass dbaas or dev or staging"
    echo "i.e. ./script dbaas"
    exit
fi
if [ -z $NAMESPACE ]; then 
    #namespace="actest"
    namespace="idcs-system"
else
    namespace=$NAMESPACE
fi

dbport=5432
if echo "$1" | grep "dbaas" -q; then
    #--- DBaas configuration
    if [ -z $DBPASS ]; then
        echo "DBaaS password not set; export DBPASS=<password>, type the password"
   read -s post_pass
    else
        post_pass=$DBPASS
    fi
    dbpostgrespod="client-1-iks-db-postgresql"
    dbname="main"
    dbsslmode="require"
    if [[ "$1" == "dbaas" ]]; then
        echo "Running against DBaaS Staging"
        dbhost="100.64.17.217"
        dbuser="psqliks_admin"
    else
        if [[ "$1" == "dbaasprod" ]]; then
            echo "Running against DBaaS Production"
            dbhost="100.64.17.217"
            dbuser="psqliks_admin"
        else
            echo DBaaS entry not found, exiting!
            exit
        fi
    fi
else
    dbpostgrespod="us-${1}-1-iks-db-postgresql-0"
    dbhost="us-${1}-1-iks-db-postgresql"
    dbuser="postgres"
    dbname="main"
    dbsslmode="disable"
    post_pass=$(kubectl -n ${namespace} get secret ${dbhost} -o jsonpath={.data.postgres-password} | base64 -d)
fi

kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c 'SELECT * FROM addoncompatibilityk8s'"

read -p "Type the addOn name ie. kube-proxy: " addonname
read -p "Type the addOn version ie. 1.27.5: " addonversion
read -p "Type the K8S version i.e. 1.27.5: " k8sversion
#addonrepo=$(echo ${addonrepo} | sed "s/\//\\\\\//g")
#echo $addonrepo
fulladdonname=$(echo ${addonname}-${addonversion})

read -p "Continue pushing changes?: " answer
if [[ "$answer" == "y" ]]; then
    fulladdonname=$(echo "'${fulladdonname}'")
    addonname=$(echo "'${addonname}-%'")
    k8sversion=$(echo "'${k8sversion}'")

    runcmd=$(echo "UPDATE addoncompatibilityk8s SET addonversion_name=${fulladdonname} WHERE k8sversion_name=${k8sversion} AND addonversion_name LIKE ${addonname}");
    kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmd}\""

    kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c 'SELECT * FROM addoncompatibilityk8s'"
fi
