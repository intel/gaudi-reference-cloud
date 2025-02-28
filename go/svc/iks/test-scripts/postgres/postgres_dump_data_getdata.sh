#! /bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

# DBaaS Vault config https://internal-placeholder.com/ui/vault/secrets/secret/show/dbaas/psql-iks/customer

if [ "$#" -ne 1 ]; then
    echo "Incorrect number of arguments, need to pass dbaas or dev or staging and the table"
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
        #dbuser="iks_user1"
    else
        if [[ "$1" == "dbaas3" ]]; then
            echo "Running against DBaaS Staging 3"
            dbhost="100.64.5.8"
            dbuser="psqliks_admin"
        else
            if [[ "$1" == "dbaasprod" ]]; then
                echo "Running against DBaaS Staging"
                dbhost="100.64.17.217"
                dbuser="psqliks_admin"
                #dbuser="iks_user1"
            else
                echo DBaaS entry not found, exiting!
                exit
	    fi
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

read -p "What environment and location is this backup for? i.e. staging-pdx: " location

rundate=$(date '+%m%d%Y-%H%M')
kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} pg_dump -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -a -t osimageinstance -t instancetype -t k8sversion -t addonversion -t k8scompatibility -t addoncompatibilityk8s -t osimageinstancecomponent > /tmp/postgres-${location}-main-new-${rundate}.sql"
kubectl -n ${namespace} cp ${dbpostgrespod}:/tmp/postgres-${location}-main-new-${rundate}.sql postgres-${location}-main-new-${rundate}.sql
kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "rm -rf /tmp/postgres*"
#INSERT INTO addoncompatibilityk8s (addonversion_name, k8sversion_name)
#UPDATE nodeprovider SET nodeprovider_name = 'Harvester' WHERE nodeprovider_name = 'caas-harvester';
#UPDATE nodeprovider SET is_default = 'true' WHERE nodeprovider_name = 'Compute';
#UPDATE nodeprovider SET is_default = 'false' WHERE nodeprovider_name = 'Harvester';
#/*RUNTIMEVERSION*/
#INSERT INTO runtimeversion (runtimeversion_name, runtime_name, version, lifecyclestate_id)
#VALUES('containerd-1.7.1','Containerd','1.7.1', 1);
#UPDATE runtimeversion SET lifecyclestate_id = 2 WHERE runtimeversion_name = 'containerd-rke2';
#INSERT INTO osimageinstancecomponent(component_name, osimageinstance_name, version, artifact_repo)
