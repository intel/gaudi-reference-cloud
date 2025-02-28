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

read -p "Set manual values? (y/n) " manualvalues

export IFS=";"
if [[ "${manualvalues}" == "y" ]]; then
    echo "Manual"
    #results=$(echo '{"items": [{"metadata": {"name": "acalvo"}, "spec": {"name": "acalvo","displayName": "Intel_ Xeon_ processors, codenamed Sapphire Rapids with high bandwidth memory (HBM) _ Flat mode", "instanceCategory": "BareMetalHost","cpu": {"cores": 56,"modelName": "4th Generation Intel_ Xeon_ Scalable Processors (Sapphire Rapids)"},"memory": {"size": "256Gi"},"disks": [{"size": "2000Gi"}]}}]}')
    read -p "Type the instance type name: " instancename
    read -p "Type the display name: " instancedisplay
    read -p "Type the instance category: " instancecat
    read -p "Type the instance family: " instancefam
    read -p "Type the instance cpu: " instancecpu
    read -p "Type the instance memory(Gi): " instancemem
    read -p "Type the instance storage(Gi): " instancestor
    results=$(echo "{\"items\":[{\"metadata\":{\"name\":\"${instancename}\"},\"spec\":{\"name\":\"${instancename}\",\"displayName\":\"${instancedisplay}\",\"instanceCategory\":\"${instancecat}\",\"cpu\":{\"cores\":${instancecpu},\"modelName\":\"${instancefam}\"},\"memory\":{\"size\":\"${instancemem}Gi\"},\"disks\":[{\"size\":\"${instancestor}Gi\"}]}}]}")
    echo $results

    echo 1
    echo $results | jq '.items[].spec | "name=\(.name)@ displayName=\(.displayName)"'
    echo 1

else
    read -p "Type the compute API URL for instance types i.e. internal-placeholder.com/compute-us-region-1-api.cloud.intel.com: " computeapiurl
    if [ -z $JWTTOKEN ]; then
        read -p "Type the jwttoken for URL: " jwttokencompute
        jwttokencompute=$jwttokencompute
    else
        jwttokencompute=$JWTTOKEN
    fi

    #name=bm-spr, displayName=4th Generation Intel® Xeon® Scalable Processors (Sapphire Rapids), cpu=56, memory=256Gi, storage=2000Gi, category=BareMetalHost, family=4th Generation Intel® Xeon® Scalable Processors (Sapphire Rapids)
    export http_proxy=http://internal-placeholder.com:911; export https_proxy=http://internal-placeholder.com:912; export no_proxy=localhost,127.0.0.0/8,10.0.0.0/8

    results=$(https_proxy=http://internal-placeholder.com:912 curl -X 'GET' "https://${computeapiurl}/v1/instancetypes" -H 'accept: application/json' -H "Authorization: Bearer ${jwttokencompute}")
    #echo $results
fi

#for i in $(https_proxy=http://internal-placeholder.com:912 curl -X 'GET' "https://${computeapiurl}/v1/instancetypes" -H 'accept: application/json' -H "Authorization: Bearer ${jwttokencompute}" | jq '.items[].spec | "name=\(.name)@ displayName=\(.displayName)@ cpu=\(.cpu.cores)@ memory=\(.memory.size)@ storage=\(.disks[].size)@ category=\(.instanceCategory)@ family=\(.cpu.modelName);"' | sed -e 's/"//g'  | tr -d "\n"); do
for i in $(echo $results | jq '.items[].spec | "name=\(.name)@ displayName=\(.displayName)@ cpu=\(.cpu.cores)@ memory=\(.memory.size)@ storage=\(.disks[].size)@ category=\(.instanceCategory)@ family=\(.cpu.modelName);"' | sed -e 's/"//g'  | tr -d "\n"); do
    if [[ "$dbuser" == "postgres" ]]; then
        unset http_proxy; unset https_proxy; unset no_proxy;
    fi
    echo $i
    instancetypename=$(echo $i | cut -d"@" -f1 | cut -d"=" -f2)
    instancetypedisplay=$(echo $i | cut -d"@" -f2 | cut -d"=" -f2)
    instancetypecpu=$(echo $i | cut -d"@" -f3 | cut -d"=" -f2)
    instancetypememory=$(echo $i | cut -d"@" -f4 | cut -d"=" -f2 | sed -e "s/Gi//g")
    instancetypestorage=$(echo $i | cut -d"@" -f5 | cut -d"=" -f2 | sed -e "s/Gi//g")
    instancetypecategory=$(echo $i | cut -d"@" -f6 | cut -d"=" -f2)
    instancetypefamily=$(echo $i | cut -d"@" -f7 | cut -d"=" -f2)
    #echo $instancetypename
    read -p "Override Family ${instancetypefamily}?: " overanswer
    if [[ "$overanswer" == "y" ]]; then
        read -p "Type new Family for ${instancetypename}: " instancetypefamily
    fi
    cp postgres_add_instancetype.sql postgres_add_instancetype-${instancetypename}.sql
    sed -i "s/TYPENAME/${instancetypename}/g" postgres_add_instancetype-${instancetypename}.sql
    sed -i "s/DISPLAYNAME/${instancetypedisplay}/g" postgres_add_instancetype-${instancetypename}.sql
#    sed -i "s/OVERRIDE/${instancetypeoverride}/g" postgres_add_instancetype-${instancetypename}.sql
    sed -i "s/MEMORY/${instancetypememory}/g" postgres_add_instancetype-${instancetypename}.sql
    sed -i "s/CPU/${instancetypecpu}/g" postgres_add_instancetype-${instancetypename}.sql
    sed -i "s/STORAGE/${instancetypestorage}/g" postgres_add_instancetype-${instancetypename}.sql
    sed -i "s/CATEGORY/${instancetypecategory}/g" postgres_add_instancetype-${instancetypename}.sql
    sed -i "s/FAMILY/${instancetypefamily}/g" postgres_add_instancetype-${instancetypename}.sql
    instancetypenameu=$(echo "'${instancetypename}'")
    runcmd="SELECT instancetype_name FROM instancetype WHERE instancetype_name = ${instancetypenameu}"
    #echo "kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmd}\""
    selecttype=$(kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmd}\"")
    if echo $selecttype | grep ${instancetypename} -q; then
        echo UPDATE
	sed -i "s/INSERTCMD.*$//g" postgres_add_instancetype-${instancetypename}.sql
	sed -i "s/UPDATECMD //g" postgres_add_instancetype-${instancetypename}.sql
    else
        echo INSERT
	sed -i "s/UPDATECMD.*$//g" postgres_add_instancetype-${instancetypename}.sql
	sed -i "s/INSERTCMD //g" postgres_add_instancetype-${instancetypename}.sql
	#cat postgres_add_instancetype-${instancetypename}.sql
    fi

    cat postgres_add_instancetype-${instancetypename}.sql 
    read -p "Continue pushing changes?(y/n - no will skip to next type): " answer
    if [[ "$answer" == "y" ]]; then
        kubectl -n ${namespace} cp ./postgres_add_instancetype-${instancetypename}.sql ${dbpostgrespod}:/tmp
        kubectl -n ${namespace} exec ${dbpostgrespod} -- chmod 666 /tmp/postgres_add_instancetype-${instancetypename}.sql
        kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -f /tmp/postgres_add_instancetype-${instancetypename}.sql"
    fi
    #cat postgres_add_instancetype-${instancetypename}.sql
    rm postgres_add_instancetype-${instancetypename}.sql
    echo "--------------------"
done
kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c 'SELECT * FROM instancetype'"
