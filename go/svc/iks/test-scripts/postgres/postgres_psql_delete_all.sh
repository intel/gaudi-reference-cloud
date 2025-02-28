# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
#dbport="5432"
#dbhost="us-dev-1-iks-db-postgresql"
#dbuser="postgres"
#dbname="main"
#dbsslmode="disable"
#dbhost="100.64.17.217"
#dbuser="psqliks_admin"
#dbname="main"
#dbsslmode="require"

export PGPASSWORD=${1}
dbhost=${2}
dbuser=${3}
dbname=${4}
dbsslmode=${5}
dbport=${6}

psql --set=sslmode=${dbsslmode} -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c "SELECT table_name FROM information_schema.tables where tables.table_catalog = '${dbname}' and tables.table_schema = 'public' and table_type = 'BASE TABLE';" | egrep -v "table|---|rows" | xargs -I {} psql -U ${dbuser} -d ${dbname} -h ${dbhost} -p ${dbport} -c "DROP TABLE {} CASCADE"
