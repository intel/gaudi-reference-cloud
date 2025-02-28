CREATE TABLE IF NOT EXISTS storageprovider (
   storageprovider_name VARCHAR(35) PRIMARY KEY,
   is_default BOOLEAN NOT NULL DEFAULT FALSE
);

INSERT INTO storageprovider (storageprovider_name, is_default)
VALUES ('vast', true);

CREATE TABLE IF NOT EXISTS storagestate (
   storagestate_name VARCHAR(15) PRIMARY KEY,
   description VARCHAR(250)
);

CREATE TABLE IF NOT EXISTS storage (
   storage_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
   storageprovider_name VARCHAR(35),
   cluster_id INT,
   kubernetes_status JSON,
   storagestate_name VARCHAR(15) NOT NULL,
   size INT NOT null,
   FOREIGN KEY (storageprovider_name)
      REFERENCES storageprovider (storageprovider_name),
   FOREIGN KEY (cluster_id)
      REFERENCES cluster(cluster_id),
   FOREIGN KEY (storagestate_name)
      REFERENCES storagestate(storagestate_name)
);


INSERT INTO storagestate (storagestate_name)
VALUES ('Active'), ('Error'), ('Updating'), ('Deleting'), ('Deleted') ON CONFLICT DO NOTHING;