{{- if .DbSeedEnabled }}
INSERT INTO defaultconfig(name, value) VALUES('cloudmonitorEnable','true');
{{- end }}