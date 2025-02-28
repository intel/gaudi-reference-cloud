#! /bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

# DBaaS Vault config https://internal-placeholder.com/ui/vault/secrets/secret/show/dbaas/psql-iks/customer

if [ "$#" -ne 1 ]; then
    echo "Incorrect number of arguments, need to pass dbaas or dev or staging and the table"
    echo "i.e. ./script dbaas datafile"
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

read -p "Add data to database? (y/n): " adddata

# Add all data to new DB
if [[ "$adddata" == "y" ]]; then
    read -p "Data file? i.e. postgres-staging-1-main-new-05302024-0856.sql: " datafile
    kubectl -n ${namespace} cp $datafile ${dbpostgrespod}:/tmp/$datafile
    runcmd=$(echo "UPDATE runtimeversion SET lifecyclestate_id=(SELECT lifecyclestate_id FROM lifecyclestate WHERE name = 'Staged')")
    kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmd}\"" 
    runcmd=$(echo "INSERT INTO runtimeversion (runtimeversion_name, runtime_name, version, lifecyclestate_id) Select 'containerd-1.7.1','Containerd','1.7.1', lifecyclestate_id from lifecyclestate where name = 'Active'")
    kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmd}\"" 
    runcmd=$(echo "INSERT INTO runtimeversion (runtimeversion_name, runtime_name, version, lifecyclestate_id) Select 'containerd-1.7.7','Containerd','1.7.7', lifecyclestate_id from lifecyclestate where name = 'Active'")
    kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmd}\"" 
    kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -f /tmp/$datafile" 
    kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "rm -rf /tmp/$datafile"
fi

# Cleanup metadata for non-active k8sversion
read -p "Clean up database? (y/n): " cleandata
if [[ "$cleandata" == "y" ]]; then
    runcmd=$(echo "delete from addoncompatibilityk8s WHERE k8sversion_name NOT IN (select k8sversion_name from k8sversion where lifecyclestate_id NOT IN (Select lifecyclestate_id from lifecyclestate where name != 'Active'))")
    kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmd}\"" 
    runcmd=$(echo "delete from k8scompatibility WHERE k8sversion_name NOT IN (select k8sversion_name from k8sversion where lifecyclestate_id NOT IN (Select lifecyclestate_id from lifecyclestate where name != 'Active'))")
    kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmd}\"" 
    runcmd=$(echo "delete from osimageinstancecomponent WHERE osimageinstance_name IN (select osimageinstance_name from osimageinstance WHERE k8sversion_name NOT IN (select k8sversion_name from k8sversion where lifecyclestate_id NOT IN (Select lifecyclestate_id from lifecyclestate where name
    kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmd}\"" 
     != 'Active')))")
    kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmd}\"" 
    runcmd=$(echo "delete from osimageinstance WHERE k8sversion_name NOT IN (select k8sversion_name from k8sversion where lifecyclestate_id NOT IN (Select lifecyclestate_id from lifecyclestate where name != 'Active'))")
    kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmd}\"" 
    runcmd=$(echo "delete from k8sversion where lifecyclestate_id != (Select lifecyclestate_id from lifecyclestate where name = 'Active');")
    kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmd}\"" 
fi

#Update default configs
read -p "Default configs? (y/n): " defultdata
if [[ "$defultdata" == "y" ]]; then
    for i in region availabilityzone vnet ilb_environment ilb_usergroup ilb_customer_environment ilb_customer_usergroup cp_cloudaccountid; do
        echo $i
        read -p "Update $i? (y/n): " updateConfig
        if [[ "$updateConfig" == "y" ]]; then
            read -p "Value for the $i key: " valuekey
	    runcmd=$(echo "UPDATE defaultconfig SET value = '${valuekey}' WHERE name = '${i}'")
            kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmd}\"" 
	fi
    done
fi












