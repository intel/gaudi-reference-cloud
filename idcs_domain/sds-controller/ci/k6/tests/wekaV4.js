// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import wekaV4 from 'k6/x/intel_internal/weka/v4';
import http from 'k6/http';
import { check } from 'k6';
import { sleep } from 'k6';

export let options = {
    iterations: 1,
    thresholds: {
        "checks": [{ threshold: "rate==1.0", abortOnFail: false }]
    }
};

export default function () {
    const wekaPort = 5000;
    wekaV4.startAPI(wekaPort, true);

    // Give server time to get up
    sleep(0.3);

    check(http.get(`http://localhost:${wekaPort}/api/v2/cluster`, {
        headers: {
            'Authorization': 'Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ',
        }
    }), {
        'Return correct status': (r) => {
            return r.status === 200
        }
    })

}