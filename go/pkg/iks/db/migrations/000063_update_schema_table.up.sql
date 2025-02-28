ALTER TABLE k8snode ADD COLUMN weka_storage_client_id VARCHAR(63);
ALTER TABLE k8snode ADD COLUMN weka_storage_status VARCHAR(63);
ALTER TABLE k8snode ADD COLUMN weka_storage_custom_status VARCHAR(63);
ALTER TABLE k8snode ADD COLUMN weka_storage_message VARCHAR(63);