#! /bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

#K8SVERSION i.e. 1.27
#K8SPATCHVERSION i.e. 1.27.5
#CPIMINAME i.e. iks-u22-cd-cp-1.27.5-23-9-18
#WKIMINAME i.e. iks-u22-cd-wk-1.27.5-23-9-18

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

read -p "Type the k8spatchversion to activate ie. 1.27.5: " k8spatchversion
k8sversion=$(echo $k8spatchversion | cut -d"." -f1,2)

cp postgres_activate_k8sversion.sql postgres_activate_k8sversion-${k8spatchversion}.sql
sed -i "s/K8SVERSION/${k8sversion}/g" postgres_activate_k8sversion-${k8spatchversion}.sql
sed -i "s/K8SPATCHVERSION/${k8spatchversion}/g" postgres_activate_k8sversion-${k8spatchversion}.sql

cat postgres_activate_k8sversion-${k8spatchversion}.sql

read -p "Continue pushing changes?: " answer
if [[ "$answer" == "y" ]]; then
    kubectl -n ${namespace} cp ./postgres_activate_k8sversion-${k8spatchversion}.sql ${dbpostgrespod}:/tmp
    kubectl -n ${namespace} exec ${dbpostgrespod} -- chmod 666 /tmp/postgres_activate_k8sversion-${k8spatchversion}.sql
    kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -f /tmp/postgres_activate_k8sversion-${k8spatchversion}.sql"
fi

rm postgres_activate_k8sversion-${k8spatchversion}.sql
