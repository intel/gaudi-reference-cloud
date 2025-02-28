// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import { check } from 'k6';
import grpc from 'k6/net/grpc';
import exec from 'k6/execution';

const client = new grpc.Client();
client.load([
    '../../../../../_main/api',
    '../../../../../googleapis~override',
    '../../../../../_main~non_module_dependencies~com_github_bufbuild_protovalidate/proto/protovalidate/buf/validate/_virtual_imports/validate_proto',
    '../../../../../_main~non_module_dependencies~com_github_bufbuild_protovalidate/proto/protovalidate/buf/validate/_virtual_imports/expression_proto',
    '../../../../../_main~non_module_dependencies~com_github_bufbuild_protovalidate/proto/protovalidate/buf/validate/priv/_virtual_imports/priv_proto',
    '../../../..'
],
    __ENV.S3_PROTO,
    __ENV.USER_PROTO,
    __ENV.VAST_FILESYSTEM_PROTO,
);

export let options = {
    scenarios: {
        default: {
            vus: 20,
            iterations: 20,
            executor: 'shared-iterations',
            maxDuration: '18000s',
        }
    },
    thresholds: {
        "checks": [{ threshold: "rate==1.0", abortOnFail: false }]
    }
};

export default function () {
    let nsIds = []

    client.connect('localhost:5000', {
        plaintext: true,
        timeout: "2s",
    });

    for (let j = 1; j < 41; j++) {
       let ipFilters = [];

        for (let f = 1; f < 6; f++) {
            for (let n = 0; n < 256; n++) {
                ipFilters.push({ "start": `10.${exec.vu.idInTest}.${f+j*f}.${f}`, "end": `10.${exec.vu.idInTest}.${f+j*f}.${f}` })
            }
        }
        
        check(client.invoke('intel.storagecontroller.v1.NamespaceService/CreateNamespace',
            {
                "cluster_id": {
                    "uuid": "00000000-0000-0000-0000-000000000000",
                },
                "name": `sds_stress_${exec.vu.idInTest}_${j}`,
                "quota": {
                    "total_bytes": 500000000
                },
                "ipFilters": ipFilters,
            }), {
            "Is ok": (r) => {
                if (r.status === grpc.StatusOK) {
                    nsIds.push(r.message.namespace.id)
                    return true
                } else {
                    console.log(r)
                    return false
                }
            }
        });
    }


    for (let nsId of nsIds) {
        check(client.invoke('intel.storagecontroller.v1.NamespaceService/DeleteNamespace',
            {
                "namespace_id": nsId,
            }), {
            'DeleteNamespace status is OK': (r) => r && r.status === grpc.StatusOK,
        });
    }

}
