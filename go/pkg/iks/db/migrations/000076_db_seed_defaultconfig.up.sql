{{- if .DbSeedEnabled }}
UPDATE defaultconfig
SET value = '{{.Ilb_environment}}'
WHERE name = 'ilb_environment';

UPDATE defaultconfig
SET value = '{{.Ilb_usergroup}}'
WHERE name = 'ilb_usergroup';

UPDATE defaultconfig
SET value = '{{.Ilb_customer_environment}}'
WHERE name = 'ilb_customer_environment';

UPDATE defaultconfig
SET value = '{{.Ilb_customer_usergroup}}'
WHERE name = 'ilb_customer_usergroup';

UPDATE defaultconfig
SET value = '{{.Availabilityzone}}'
WHERE name = 'availabilityzone';

UPDATE defaultconfig
SET value = '{{.Vnet}}'
WHERE name = 'vnet';

UPDATE defaultconfig
SET value = '{{.Cp_cloudaccountid}}'
WHERE name = 'cp_cloudaccountid';

UPDATE defaultconfig
SET value = '{{.Region}}'
WHERE name = 'region'
{{- end }}