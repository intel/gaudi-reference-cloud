# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
source ./go/svc/dpai/test-scripts/common.sh

forwardPort us-dev-1-dpai-db-postgresql 5432 5432

PG_PASSWORD=$(kubectl get secret us-dev-1-dpai-db-postgresql -n idcs-system -o jsonpath='{.data.postgres-password}' | base64 --decode)
docker run --rm -e "PG_PASSWORD=$PG_PASSWORD" --network host -v $(pwd)/go/pkg/dpai/db:/src -w /src sqlc/sqlc generate 