import { SharedArray } from "k6/data";

export var data = new SharedArray("payload", function () {
  const payload = JSON.parse(
    open("../test_config/prod_catalog_load_users.json")
  );
  return payload;
});

export var configData = JSON.parse(
  open("../test_config/prod_catalog_load_config.json")
);
