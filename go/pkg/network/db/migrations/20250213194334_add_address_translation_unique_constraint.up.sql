-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
CREATE UNIQUE INDEX unique_address_translation_idx ON address_translation (
                                                                           cloud_account_id,
                                                                           (value->'spec'->>'portId'),
                                                                           (value->'spec'->>'translationType')
                                                                          );