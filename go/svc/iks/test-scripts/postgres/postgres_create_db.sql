-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
SELECT 'CREATE DATABASE main' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'main')\gexec
