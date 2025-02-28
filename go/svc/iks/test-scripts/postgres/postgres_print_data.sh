#! /bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

# DBaaS Vault config https://internal-placeholder.com/ui/vault/secrets/secret/show/dbaas/psql-iks/customer

if [ "$#" -lt 2 ]; then
    echo "Incorrect number of arguments, need to pass dbaas or dev or staging and the table"
    echo "i.e. ./script dbaas k8sversion"
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

tablename=${2}

if [ ! ${3} ]; then
    fields="*"
else
    fields=${3}
fi

if [ ! ${4} ]; then
    whereclause=""
else
    whereclause=$(echo "WHERE ${4}") # This should be field=\'value\', i.e. k8sversion_name=\'1.27.8\'
fi

echo "kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c 'SELECT ${fields} FROM ${tablename} LIMIT 1'"
tablenames=$(kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c 'SELECT ${fields} FROM ${tablename} LIMIT 1'")

runcmd="SELECT ${fields} FROM ${tablename} ${whereclause}"
if echo ${tablenames} | grep created_date -q; then
#    kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c 'SELECT ${fields} FROM ${tablename} ${whereclause} ORDER BY created_date'"
    runcmd=$(echo "${runcmd} ORDER BY created_date")
fi



#echo "kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmd}\""
kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmd}\""
