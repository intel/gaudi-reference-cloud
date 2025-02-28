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

imioverride="'false'"

read -p "Update for instancetype or imi?: " answer
# Do the overrides for Instance Type
if [[ "$answer" == "instancetype" ]]; then
    kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c 'SELECT * FROM instancetype'"
    read -p "Type the instance type name i.e. vm-spr-tny: " instancetypename
    instancetypename=$(echo "'${instancetypename}'")
    runcmd=($echo "SELECT instancetype_name, instancecategory, instancetypefamiliy, imi_override FROM instancetype WHERE instancetype_name=${instancetypename}")
    instanceinfo=$(kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql -t --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmd}\"")
    echo "----------------"
    echo ${instanceinfo}
    instancecatq=$(echo $instanceinfo | cut -d"|" -f2 | sed -e "s/^ //g" | sed -e "s/ $//g")
    instancefamq=$(echo $instanceinfo | cut -d"|" -f3 | sed -e "s/^ //g" | sed -e "s/ $//g")
    echo "Instance type has Cat: '${instancecatq}' Fam: '${instancefamq}'"
    echo "----------------"
    instancecat=$(echo "'${instancecatq}'")
    instancefam=$(echo "'${instancefamq}'")
    #runcmd=$(echo "SELECT osimageinstance_name, k8sversion_name, osimage_name, runtime_name, provider_name, instancetypecategory, created_date FROM osimageinstance WHERE instancetypecategory=${instancecat} AND instancetypefamiliy=${instancefam} AND lifecyclestate_id=(SELECT lifecyclestate_id FROM lifecyclestate WHERE name='Active' ORDER BY k8sversion_name, created_date DESC)")
    runcmd=$(echo "SELECT osimageinstance_name, k8sversion_name, osimage_name, runtime_name, provider_name, instancetypecategory, created_date FROM osimageinstance WHERE instancetypecategory=${instancecat} AND instancetypefamiliy=${instancefam} AND lifecyclestate_id=(SELECT lifecyclestate_id FROM lifecyclestate WHERE name='Active') AND (k8sversion_name IN (SELECT DISTINCT k8sversion_name from k8sversion where lifecyclestate_id=1) OR k8sversion_name IN (select DISTINCT k8sversion_name from nodegroup)) ORDER BY k8sversion_name, created_date DESC")
    #echo $runcmd
    for i in $(kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmd}\"" | grep -i ${instancecatq} | sed -e "s/ //g"); do
        echo OS Image with Category and Family: $i
	iminame=$(echo $i | cut -d "|" -f1)
        newoverride="\"${instancetypename}\": \"${iminame}\""
	k8sversion=$(echo $i | cut -d "|" -f2 | xargs -I {} echo "'{}'")
	osname=$(echo $i | cut -d "|" -f3 | xargs -I {} echo "'{}'")
	runtimename=$(echo $i | cut -d "|" -f4 | xargs -I {} echo "'{}'")
	providername=$(echo $i | cut -d "|" -f5 | xargs -I {} echo "'{}'")
	iminame=$(echo "'${iminame}'")
	runcmdgetCP=$(echo "SELECT cp_osimageinstance_name FROM k8scompatibility WHERE k8sversion_name=${k8sversion} and osimage_name=${osname} and runtime_name=${runtimename} and provider_name=${providername} LIMIT 1")
	resultsCP=$(kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql -t --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmdgetCP}\"" | sed -e "s/^ //")
	resultsCP=$(echo "'${resultsCP}'")
	echo "SELECT k8sversion_name FROM k8scompatibility WHERE k8sversion_name=${k8sversion} and osimage_name=${osname} and runtime_name=${runtimename} and provider_name=${providername} and instancetype_name=${instancetypename}"
	runcmdget=$(echo "SELECT k8sversion_name FROM k8scompatibility WHERE k8sversion_name=${k8sversion} and osimage_name=${osname} and runtime_name=${runtimename} and provider_name=${providername} and instancetype_name=${instancetypename}")
	results=$(kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql -t --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmdget}\"")
	echo $results
	if  [[ -z $results ]]; then
	    #echo $results
            echo K8sversion: ${k8sversion} 

            read -p "Insert new k8sversion/instancetype (y/n)?: " answeroverride
            if [[ "$answeroverride" == "y" ]]; then
                runcmdupdcmp=$(echo "INSERT INTO k8scompatibility(k8sversion_name, osimage_name, runtime_name, provider_name, instancetype_name, cp_osimageinstance_name, wrk_osimageinstance_name, lifecyclestate_id) SELECT ${k8sversion}, ${osname}, ${runtimename}, ${providername}, ${instancetypename},${resultsCP},${iminame}, lifecyclestate_id FROM lifecyclestate WHERE name = 'Active'")
                kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmdupdcmp}\"" 
            else
                echo Not insert
            fi
        else
            echo "Instance type already in k8s for k8sversion"
	fi
	echo "----------------"
	#override=$(echo "'$override'" | sed -e 's/"/\\"/g')
    done
    #instancetypename=$(echo "'$instancetypename'")
    #echo "Setting Instance Type Override to 'true'"
    #runcmdupdtype=$(echo "UPDATE instancetype SET imi_override='true' WHERE instancetype_name=${instancetypename}")
    #kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmdupdtype}\"" 
else
    if [[ "$answer" == "imi" ]]; then
        #Do the overrides for os image instance 
        kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c 'SELECT osimageinstance_name, k8sversion_name, osimage_name, runtimeversion_name, created_date, nodegrouptype_name, lifecyclestate_id AS Stateid, provider_name, instancetypecategory, instancetypefamiliy FROM osimageinstance'" 
        read -p "Type the imi name i.e. iks-u22-cd-wk-1-27-4-23-09-19: " iminame
        imiinfo=$(kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c 'SELECT osimageinstance_name, instancetypecategory, instancetypefamiliy, k8sversion_name, osimage_name, runtime_name, provider_name FROM osimageinstance'" | grep ${iminame})
        echo "----------------"
	echo ${imiinfo}
	imiinfoNoSpace=$(echo ${imiinfo} | sed -e "s/ //g")
        imicatq=$(echo $imiinfo | cut -d"|" -f2 | sed -e "s/^ //g" | sed -e "s/ $//g")
        imifamq=$(echo $imiinfo | cut -d"|" -f3 | sed -e "s/^ //g" | sed -e "s/ $//g")
        k8sversion=$(echo $imiinfoNoSpace | cut -d "|" -f4 | xargs -I {} echo "'{}'")
        osname=$(echo $imiinfoNoSpace | cut -d "|" -f5 | xargs -I {} echo "'{}'")
        runtimename=$(echo $imiinfoNoSpace | cut -d "|" -f6 | xargs -I {} echo "'{}'")
        providername=$(echo $imiinfoNoSpace | cut -d "|" -f7 | xargs -I {} echo "'{}'")
        echo "IMI has Cat: '${imicatq}' Fam: '${imifamq}'"
        echo "----------------"
        imicat=$(echo "'${imicatq}'")
        imifam=$(echo "'${imifamq}'")
        iminame=$(echo "'${iminame}'")
        runcmd=$(echo "SELECT instancetype_name, instancecategory, instancetypefamiliy FROM instancetype WHERE instancecategory=${imicat} AND instancetypefamiliy=${imifam}")
        for i in $(kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmd}\"" | grep -i ${imicatq} | sed -e "s/ //g"); do
            echo Instance Type with Category and Family: $i
            instancetypename=$(echo $i | cut -d "|" -f1)
            instancetypename=$(echo "'${instancetypename}'")

            runcmdget=$(echo "SELECT wrk_osimageinstance_name FROM k8scompatibility WHERE k8sversion_name=${k8sversion} and osimage_name=${osname} and runtime_name=${runtimename} and provider_name=${providername} and instancetype_name=${instancetypename}")
	    results=$(kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql -t --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmdget}\"")
            #echo 1 $results $iminame
            if  [ ! -z $results ]; then
                echo UPDATE
                runcmdupdcmp=$(echo "UPDATE k8scompatibility SET wrk_osimageinstance_name=${iminame} WHERE k8sversion_name=${k8sversion} and osimage_name=${osname} and runtime_name=${runtimename} and provider_name=${providername} and instancetype_name=${instancetypename}")
	    else
                echo INSERT
                runcmdupdcmp=$(echo "INSERT INTO k8scompatibility SET wrk_osimageinstance_name=${iminame} WHERE k8sversion_name=${k8sversion} and osimage_name=${osname} and runtime_name=${runtimename} and provider_name=${providername} and instancetype_name=${instancetypename}")
            fi
	    echo $runcmdupdcmp
            echo K8sversion: ${k8sversion}

            read -p "Insert/Update for k8sversion (y/n)?: " answeroverride
            if [[ "$answeroverride" == "y" ]]; then
                #runcmdupdcmp=$(echo "UPDATE k8scompatibility SET instancetype_wk_override=${override} WHERE k8sversion_name=${k8sversion} and osimage_name=${osname} and runtime_name=${runtimename} and provider_name=${providername}")
                kubectl -n ${namespace} exec ${dbpostgrespod} -- bash -c "PGPASSWORD=${post_pass} psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c \"${runcmdupdcmp}\"" 
            else
	        echo Change not applied
            fi
	    echo "----------------"
        done
    fi
fi
