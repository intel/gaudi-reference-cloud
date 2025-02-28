#!/bin/bash
k6 run --out json=ramping.js ./GetCloudAccountBenchmark.js
k6 run --out json=ramping.js ./GetCreditsBenchmark.js
k6 run --out json=ramping.js ./GetInvoicesBenchmark.js
k6 run --out json=ramping.js ./GetMembersBenchmark.js
k6 run --out json=ramping.js ./GetProductsBenchmark.js
k6 run --out json=ramping.js ./GetUsagesBenchmark.js
