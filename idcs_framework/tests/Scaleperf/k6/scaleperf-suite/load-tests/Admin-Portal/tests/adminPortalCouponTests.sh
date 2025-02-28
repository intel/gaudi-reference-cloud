#!/bin/bash
k6 run --out json=coupons_constant.js ./GetCouponsTest-Constant.js
k6 run --out json=coupons_pervu.js ./GetCouponsTest-PerVu.js
k6 run --out json=coupons_ramping.js ./GetCouponsTest-Ramping.js
k6 run --out json=coupons_shared.js ./GetCouponsTest-Shared.js