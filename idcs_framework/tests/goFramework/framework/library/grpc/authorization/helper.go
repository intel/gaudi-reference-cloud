package authorization

type CouponPayload struct {
	Error     string `json:"error"`
	FamilyId  string `json:"familyId"`
	ProductId string `json:"productId"`
	Status    string `json:"status"`
	VendorId  string `json:"vendorId"`
}

var GET_COUPONS_ENDPOINT = "proto.BillingCouponService/Read"

var CREATE_COUPON_ENDPOINT = "proto.BillingCouponService/Create"
