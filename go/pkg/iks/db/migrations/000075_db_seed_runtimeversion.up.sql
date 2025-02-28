{{- if .DbSeedEnabled }}
-- Update all the existing runtime versions to 'Staged' state
UPDATE runtimeversion
SET lifecyclestate_id=(SELECT lifecyclestate_id FROM lifecyclestate WHERE name = 'Staged');

-- Add new runtime version and set the lifecycle state to Active
INSERT INTO runtimeversion (runtimeversion_name, runtime_name, version, lifecyclestate_id)
Select 'containerd-1.7.1', 'Containerd', '1.7.1', lifecyclestate_id
from lifecyclestate
where name = 'Active';

-- Add new runtime version and set the lifecycle state to Active
INSERT INTO runtimeversion (runtimeversion_name, runtime_name, version, lifecyclestate_id)
Select 'containerd-1.7.7', 'Containerd', '1.7.7', lifecyclestate_id
from lifecyclestate
where name = 'Active'
{{- end }}