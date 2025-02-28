#! /bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

#K8SVERSION i.e. 1.27
#K8SPATCHVERSION i.e. 1.27.5
#CPIMINAME i.e. iks-u22-cd-cp-1.27.5-23-9-18
#WKIMINAME i.e. iks-u22-cd-wk-1.27.5-23-9-18

# DBaaS Vault config https://internal-placeholder.com/ui/vault/secrets/secret/show/dbaas/psql-iks/customer

if [ "$#" -lt 1 ]; then
    echo "Incorrect number of arguments, need to pass dbaas or dev or staging"
    echo "i.e. ./script dbaas"
    exit
else
    if [ "$#" -eq 9 ]; then
        noread=true
        k8spatchversion=$2
        cpiminame=$3
        containerdname=$4
        addonproxy=$5
        addoncoredns=$6
        addoncalicoop=$7
        addoncalicoconf=$8
        addonkonnectivity=$9
    else
        noread=false
    fi
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

if [ "${noread}" == "false" ]; then
    read -p "Type the k8spatchversion ie. 1.27.5: " k8spatchversion
    read -p "Type the CP IMI (all lowercase) ie. iks-u22-cd-cp-1.27.5-23-9-18: " cpiminame
    read -p "Type the Containerd version name ie. containerd-1.7.7: " containerdname
    read -p "Type the kube-proxy addon version name ie. kube-proxy-1.27.1: " addonproxy
    read -p "Type the coredns addon version name ie. coredns-1.7.1: " addoncoredns
    read -p "Type the calico-operator addon version name ie. calico-operator-3.26.0: " addoncalicoop
    read -p "Type the calico-config addon version name ie. calico-config-3.26.0: " addoncalicoconf
    read -p "Type the konnectivity-agent addon version name ie. konnectivity-agent-0.0.0: " addonkonnectivity
fi

k8sversion=$(echo $k8spatchversion | cut -d"." -f1,2)
cp postgres_add_k8sversion.sql postgres_add_k8sversion-${k8spatchversion}.sql
sed -i "s/K8SVERSION/${k8sversion}/g" postgres_add_k8sversion-${k8spatchversion}.sql
sed -i "s/K8SPATCHVERSION/${k8spatchversion}/g" postgres_add_k8sversion-${k8spatchversion}.sql
sed -i "s/CPIMINAME/${cpiminame}/g" postgres_add_k8sversion-${k8spatchversion}.sql
sed -i "s/CONTAINERDNAME/${containerdname}/g" postgres_add_k8sversion-${k8spatchversion}.sql
sed -i "s/ADDONPROXY/${addonproxy}/g" postgres_add_k8sversion-${k8spatchversion}.sql
sed -i "s/ADDONCOREDNS/${addoncoredns}/g" postgres_add_k8sversion-${k8spatchversion}.sql
sed -i "s/ADDONCALOPER/${addoncalicoop}/g" postgres_add_k8sversion-${k8spatchversion}.sql
sed -i "s/ADDONCALCONF/${addoncalicoconf}/g" postgres_add_k8sversion-${k8spatchversion}.sql
sed -i "s/ADDONKONN/${addonkonnectivity}/g" postgres_add_k8sversion-${k8spatchversion}.sql

k8spatchversionq="'${k8spatchversion}'"
newcol=""
runcmdget=$(echo "SELECT osimageinstance_name, instancetypecategory, instancetypefamiliy FROM osimageinstance WHERE k8sversion_name=${k8spatchversionq} AND nodegrouptype_name='worker'")
for i in $(kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql -t --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmdget}\"" | sed -e "s/ /@@/g"); do
    inst=$(echo $i | cut -d "|" -f1 | sed -e "s/@@//g" | xargs -I {} echo "'{}'")
    instcat=$(echo $i | cut -d "|" -f2 | sed -e "s/@@//g" | xargs -I {} echo "'{}'")
    instfam=$(echo $i | cut -d "|" -f3 | sed -e "s/@@/ /g" | xargs -I {} echo "'{}'")
    echo IMI: $inst Cat: $instcat Family: $instfam
    runcmdget=$(echo "SELECT instancetype_name FROM instancetype WHERE instancecategory=${instcat} AND instancetypefamiliy=${instfam}")
    for j in $(kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql -t --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmdget}\""); do
        echo InstanceType: $j
	read -p "Add this instancetype/IMI? (y/n): " inimanswer
	if [ "$inimanswer" == "y" ]; then
            newcol=$(echo "INSERT INTO k8scompatibility(provider_name, runtime_name, k8sversion_name, osimage_name, cp_osimageinstance_name, wrk_osimageinstance_name, instancetype_name, lifecyclestate_id)\nSELECT 'iks','Containerd','${k8spatchversion}','Ubuntu-22-04','${cpiminame}',${inst}, '${j}',lifecyclestate_id from lifecyclestate where name = 'Active';\n${newcol}")
	fi
    done
done
#echo $newcol
sed -i "s/WKSELECT/${newcol}/g" postgres_add_k8sversion-${k8spatchversion}.sql

cat postgres_add_k8sversion-${k8spatchversion}.sql

read -p "Continue pushing changes?: " answer
if [[ "$answer" == "y" ]]; then
    if [ -z $DBPASS ]; then
        post_pass=$(kubectl -n ${namespace} get secret ${dbhost} -o jsonpath={.data.postgres-password} | base64 -d)
    else
        post_pass=$DBPASS
    fi
    kubectl -n ${namespace} cp ./postgres_add_k8sversion-${k8spatchversion}.sql ${dbpostgrespod}:/tmp
    kubectl -n ${namespace} exec ${dbpostgrespod} -- chmod 666 /tmp/postgres_add_k8sversion-${k8spatchversion}.sql
    kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -f /tmp/postgres_add_k8sversion-${k8spatchversion}.sql"
fi

rm postgres_add_k8sversion-${k8spatchversion}.sql
