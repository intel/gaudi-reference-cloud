-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
SELECT instancetype_name, displayname, lifecyclestate_id, imi_override, memory, cpu, nodeprovider_name, storage, instancecategory, instancetypefamiliy FROM instancetype;

INSERTCMD INSERT into instancetype(instancetype_name, displayname, imi_override, memory, cpu, nodeprovider_name, storage, instancecategory, instancetypefamiliy,lifecyclestate_id)
INSERTCMD Select 'TYPENAME','DISPLAYNAME','false','MEMORY','CPU','Compute','STORAGE', 'CATEGORY', 'FAMILY',lifecyclestate_id from lifecyclestate where name = 'Staged';

UPDATECMD UPDATE instancetype SET displayname='DISPLAYNAME', memory='MEMORY', cpu='CPU', storage='STORAGE', instancecategory='CATEGORY', instancetypefamiliy='FAMILY' WHERE instancetype_name = 'TYPENAME';

SELECT instancetype_name, displayname, lifecyclestate_id, imi_override, memory, cpu, nodeprovider_name, storage, instancecategory, instancetypefamiliy FROM instancetype;
