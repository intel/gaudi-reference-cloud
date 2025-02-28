# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
dbhost=${2}
dbuser=${3}
dbname=${4}
dbsslmode=${5}
dbport=${6}

export PGPASSWORD=${1}

tables=(addoncompatibilityk8s addonversion instancetype k8scompatibility osimageinstancecomponent osimageinstance k8sversion provisioninglog)

if [ d${7} == "dcluster" ]; then
    echo cluster
    psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c "DELETE FROM cluster CASCADE"
fi
for i in ${tables[@]}; do
    echo $i
    psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c "DELETE FROM ${i}"
done
