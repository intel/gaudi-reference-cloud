#!/bin/bash
k6 run --out json=constant.js ./GetProductsTest-Constant.js
k6 run --out json=pervu.js ./GetProductsTest-PerVu.js
k6 run --out json=ramping.js ./GetProductsTest-Ramping.js
k6 run --out json=shared.js ./GetProductsTest-Shared.js
k6 run --out json=vendors_constant.js ./GetVendorsTest-Constant.js
k6 run --out json=vendors_pervu.js ./GetVendorsTest-PerVu.js
k6 run --out json=vendors_ramping.js ./GetVendorsTest-Ramping.js
k6 run --out json=vendors_shared.js ./GetVendorsTest-Shared.js