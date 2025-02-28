# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
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
    dbhost="us-${1}-1-iks-db-postgresql"
    post_pass=$(kubectl -n ${namespace} get secret ${dbhost} -o jsonpath={.data.postgres-password} | base64 -d)
fi

kubectl -n ${namespace} apply -f go-migate-depl.yaml
sleep 15
dbpostgrespod=$(kubectl -n ${namespace} get po | grep  iksgomigrate | awk '{print $1}')
echo $dbpostgrespod
kubectl -n ${namespace} exec ${dbpostgrespod} -- mkdir /tmp/migrations /tmp/gorun
kubectl -n ${namespace} cp ./postgres_deploy_migrations.go ${dbpostgrespod}:/tmp/gorun/postgres_deploy_migrations.go
for i in $(ls ../../../../pkg/iks/db/migrations); do
    echo Copying migration script ${i}
    kubectl -n ${namespace} cp ../../../../pkg/iks/db/migrations/${i} ${dbpostgrespod}:/tmp/migrations/${i}
done

echo cd /tmp/gorun > gorun-exec.sh
echo go mod init gorun >> gorun-exec.sh

read -p "Set proxies for Internal? (y/n): " setproxies
if [[ "$setproxies" == "y" ]]; then
    echo "export http_proxy=http://internal-placeholder.com:911; export https_proxy=http://internal-placeholder.com:912; export no_proxy=localhost,127.0.0.0/8,10.0.0.0/8,.intel.com" >> gorun-exec.sh
fi
echo go mod tidy >> gorun-exec.sh
echo "export post_pass=${post_pass}; go run /tmp/gorun/postgres_deploy_migrations.go ${1} ${dbhost}" >> gorun-exec.sh
chmod 755 gorun-exec.sh
kubectl -n ${namespace} cp ./gorun-exec.sh ${dbpostgrespod}:/tmp/gorun-exec.sh

kubectl -n ${namespace} exec ${dbpostgrespod} -- chmod 666 /tmp/gorun/postgres_deploy_migrations.go
kubectl -n ${namespace} exec ${dbpostgrespod} -- sh /tmp/gorun-exec.sh
sleep 5
kubectl -n ${namespace} delete -f go-migate-depl.yaml
rm gorun-exec.sh
