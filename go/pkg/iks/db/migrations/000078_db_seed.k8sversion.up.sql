{{- if .DbSeedEnabled }}
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.27.1', 2, '1.27', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.27.0', 2, '1.27', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.26.4', 2, '1.26', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.26.3', 2, '1.26', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.25.8', 2, '1.25', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('v1.27.1+rke2r1', 2, '1.27', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('v1.26.4+rke2r1', 2, '1.26', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('v1.26.3+rke2r1', 2, '1.26', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.27.5', 3, '1.27', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.26.9', 2, '1.26', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.25.9', 2, '1.25', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.27.7', 2, '1.27', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.27.6', 3, '1.27', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('v1.27.2+rke2r1', 2, '1.27', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.27.4', 2, '1.27', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.27.8', 2, '1.27', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.28.3', 2, '1.28', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.28.4', 2, '1.28', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.26.13', 3, '1.26', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.24.17', 2, '1.24', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.25.16', 2, '1.25', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.29.2', 2, '1.29', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.27.9', 2, '1.27', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.27.11', 2, '1.27', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.28.5', 2, '1.28', '1', false);
INSERT INTO public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.28.7', 2, '1.28', '1', false);
INSERT into public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.30.3', 1, '1.30', '1', false);
INSERT into public.k8sversion (k8sversion_name, lifecyclestate_id, minor_version, major_version, test_version) VALUES ('1.31.3', 1, '1.31', '1', false);
{{- end }}