#!/bin/bash
k6 run --out json=cloudaccounts_constant.js ./GetCloudAccountsTest-Constant.js
k6 run --out json=cloudaccounts_pervu.js ./GetCloudAccountsTest-PerVu.js
k6 run --out json=cloudaccounts_ramping.js ./GetCloudAccountsTest-Ramping.js
k6 run --out json=cloudaccounts_shared.js ./GetCloudAccountsTest-Shared.js