-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
ALTER TABLE osinstanceimage RENAME COLUMN osinstanceimage_name TO osimageinstance_name;
ALTER TABLE osinstanceimage RENAME TO osimageinstance;
ALTER TABLE runtimecompatibilityk8s RENAME COLUMN cp_osinstanceimage_name TO cp_osimageinstance_name;
ALTER TABLE runtimecompatibilityk8s RENAME COLUMN wrk_osinstanceimage_name TO wrk_osimageinstance_name;

ALTER TABLE runtimecompatibilityk8s RENAME CONSTRAINT runtimecompatibilityk8s_cp_osinstanceimage_name_fkey TO runtimecompatibilityk8s_cp_osimageinstnace_name_fkey;
ALTER TABLE runtimecompatibilityk8s RENAME CONSTRAINT runtimecompatibilityk8s_wrk_osinstanceimage_name_fkey TO runtimecompatibilityk8s_wrk_osimageinstnace_name_fkey;

ALTER TABLE osimageinstance ADD COLUMN runtime_name VARCHAR(15);
UPDATE osimageinstance SET runtime_name='Containerd';
ALTER TABLE osimageinstance ALTER COLUMN runtime_name SET NOT NULL;
ALTER TABLE osimageinstance ADD CONSTRAINT osimageinstance_runtime_name_fkey FOREIGN KEY (runtime_name) REFERENCES runtime (runtime_name);


CREATE TABLE IF NOT EXISTS k8scompatibility (
   runtime_name VARCHAR(15) NOT NULL,
   k8sversion_name VARCHAR(63) NOT NULL,
   osimage_name VARCHAR(63) NOT NULL,
   cp_osimageinstance_name VARCHAR(63) NOT NULL,
   wrk_osimageinstance_name VARCHAR(63) NOT NULL,
   PRIMARY KEY (runtime_name, k8sversion_name, osimage_name),
   FOREIGN KEY (runtime_name)
       REFERENCES runtime (runtime_name),
   FOREIGN KEY (osimage_name)
       REFERENCES osimage (osimage_name),
   FOREIGN KEY (k8sversion_name)
       REFERENCES k8sversion (k8sversion_name)
);

INSERT INTO k8scompatibility (runtime_name, k8sversion_name, osimage_name, cp_osimageinstance_name, wrk_osimageinstance_name)
SELECT runtime_name, k8sversion_name, 'ubuntu2204','ubuntu2204-k1.22.17-c1.0.0','ubuntu2204-k1.22.17-c1.0.0' FROM runtimecompatibilityk8s;

DROP TABLE IF EXISTS runtimecompatibilityk8s;
