#!/usr/bin/env python3

import argparse
import glob
import json
import jwcrypto
import logging
import python_jwt


def main():
    parser = argparse.ArgumentParser(
        description="Convert Java Web Keys from Kubernetes to Vault JSON/PEM format which can be used by the vault client"
    )
    parser.add_argument("--input-dir", default="local/secrets/vault-jwk-validation-public-keys")
    parser.add_argument("--output-file", default="local/secrets/vault-jwt-validation-public-keys.json")
    parser.add_argument("--log-level", type=int, default=logging.INFO, help="10=DEBUG,20=INFO")
    args = parser.parse_args()

    logging.basicConfig(level=args.log_level)
    logging.debug("args=%s" % str(args))

    pem_list = []
    for input_file_name in glob.glob(args.input_dir + "/*"):
        logging.info("Reading " + input_file_name)
        with open(input_file_name, "r") as input_file:
            jwk_keys_json = input_file.read()
            jwk_keys = json.loads(jwk_keys_json)
            for jwk_key in jwk_keys["keys"]:
                jwk_json = json.dumps(jwk_key)
                jwk = jwcrypto.jwk.JWK.from_json(jwk_json)
                pem = jwk.export_to_pem().decode("utf-8")
                pem_list = pem_list + [pem]

    output_json = json.dumps(dict(jwt_validation_pubkeys=pem_list))
    with open(args.output_file, "w") as output_file:
        output_file.write(output_json)
    logging.info("Wrote " + args.output_file)


if __name__ == "__main__":
    main()
