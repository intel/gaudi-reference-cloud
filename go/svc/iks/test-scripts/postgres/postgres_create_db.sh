#! /bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

# DBaaS Vault config https://internal-placeholder.com/ui/vault/secrets/secret/show/dbaas/psql-iks/customer

if [ "$#" -lt 1 ]; then
    echo "Incorrect number of arguments, need to pass dbaas or dev or staging"
    echo "i.e. ./script dbaas"
    exit
fi

#dbpostgrespod="us-staging-1-iks-db-postgresql-0"
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
    dbname="postgres"
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
    dbhost="us-${1}-1-iks-db-postgresql"
    dbuser="postgres"
    dbname="postgres"
    dbsslmode="disable"
    dbpostgrespod="us-${1}-1-iks-db-postgresql-0"
fi

if [ -z $NAMESPACE ]; then 
    #namespace="actest"
    namespace="idcs-system"
else
    namespace=$NAMESPACE
fi

if [ -z $post_pass ]; then
    post_pass=$(kubectl -n ${namespace} get secret ${dbhost} -o jsonpath={.data.postgres-password} | base64 -d)
fi

read -p "Continue pushing changes?: " answer
if [[ "$answer" == "y" ]]; then
    kubectl -n ${namespace} cp ./postgres_create_db.sql ${dbpostgrespod}:/tmp
    kubectl -n ${namespace} exec ${dbpostgrespod} -- chmod 666 /tmp/postgres_create_db.sql
    echo "kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c PGPASSWORD=${post_pass} psql -U ${dbuser} -d ${dbname} -h ${dbhost} --set=sslmode=required -f /tmp/postgres_create_db.sql"
    kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql -U ${dbuser} -d ${dbname} -h ${dbhost} --set=sslmode=required -f /tmp/postgres_create_db.sql"
fi
