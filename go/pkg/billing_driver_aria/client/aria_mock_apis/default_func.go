// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock

import (
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
	"github.com/pborman/uuid"
)

var AddressMap = map[string]string{
	"PaymentOption":    "Methods",
	"PaymentTermsType": "Net Terms",
	"FirstName":        "John",
	"LastName":         "Doe",
	"CompanyName":      "Company ABC",
	"City":             "AUSTIN",
	"StateProv":        "TX",
	"Country":          "US",
	"PostalCd":         "73301",
	"Email":            "test2@testmail.com",
}

func GetDefaultBillingError() []data.BillingError {
	return []data.BillingError{
		{
			InvoicingErrorCode:   0,
			InvoicingErrorMsg:    "OK",
			CollectionErrorCode:  -1,
			CollectionErrorMsg:   "not attempted",
			StatementErrorCode:   18001,
			StatementErrorMsg:    "could not create statement",
			BillingGroupNo:       billingGroupNos[len(billingGroupNos)-1],
			ClientBillingGroupId: strcnv(billingGroupNos[len(billingGroupNos)-1]),
		},
	}
}

func GetDefaultMasterPlanAssigned() []data.MasterPlansAssigned {
	return []data.MasterPlansAssigned{
		{
			PlanInstanceNo:       planInstanceNos[len(planInstanceNos)-1],
			ClientPlanInstanceId: uuid.New()[:55],
		},
	}
}

func GetDefaultInvoiceInfo() []data.InvoiceInfo {
	return []data.InvoiceInfo{
		{
			InvoiceNo:            randomDigitGen(10), //not a constant
			BillingGroupNo:       billingGroupNos[len(billingGroupNos)-1],
			ClientBillingGroupId: strcnv(billingGroupNos[len(billingGroupNos)-1]),
		},
	}
}

func GetDefaultChiefAcctInfo() []data.ChiefAcctInfo {
	return []data.ChiefAcctInfo{
		{
			ChiefAcctNo:       randomDigitGen(8),
			ChiefAcctUserId:   StringGen(8),
			ChiefClientAcctId: StringGen(55),
		},
	}
}

func GetDefaultStatementContactDetails() []data.StatementContactDetail {
	return []data.StatementContactDetail{
		{
			BillingGroupNo: 953361,
		},
	}
}

func GetDefaultBillingGroupsInfo() []data.BillingGroupsInfo {
	return []data.BillingGroupsInfo{
		{
			BillingGroupNo:       billingGroupNos[len(billingGroupNos)-1],
			ClientBillingGroupId: strcnv(billingGroupNos[len(billingGroupNos)-1]),
			NotifyMethod:         10,
			NotifyTemplateGroup:  2472,
			PaymentOption:        AddressMap["PaymentOption"],
			PaymentTermsType:     AddressMap["PaymentTermsType"],
			StmtFirstName:        AddressMap["FirstName"],
			StmtLastName:         AddressMap["LastName"],
			StmtCompanyName:      AddressMap["CompanyName"],
			StmtAddress1:         StringGen(20),
			StmtCity:             AddressMap["City"],
			StmtStateProv:        AddressMap["StateProv"],
			StmtCountry:          AddressMap["Country"],
			StmtPostalCd:         AddressMap["PostalCd"],
			StmtEmail:            AddressMap["Email"],
			StmtCellPhone:        strcnv(randomDigitGen(8)),
		},
	}
}

func GetDefaultFunctionalAcctGroup() []data.FunctionalAcctGroup {
	return []data.FunctionalAcctGroup{
		{FunctionalAcctGroupNo: randomDigitGen(8),
			ClientFunctionalAcctGroupId: "LE036"},
	}
}

func GetDefaultBillingContactInfo() []data.BillingContactInfo {
	return []data.BillingContactInfo{
		{
			PaymentMethodNo:  2,
			BillingContactNo: 0,
		},
	}
}

func GetDefaultAccountPaymentMethods() []data.AccountPaymentMethods {
	return []data.AccountPaymentMethods{
		{
			BillFirstName:         AddressMap["FirstName"],
			BillMiddelInitial:     uuid.New(),
			BillLastName:          AddressMap["LastName"],
			BillComapanyName:      AddressMap["CompanyName"],
			BillAddress1:          uuid.New(),
			BillAddress2:          uuid.New(),
			BillAddress3:          uuid.New(),
			BillCity:              AddressMap["City"],
			BillLocality:          uuid.New(),
			BillCountry:           AddressMap["Country"],
			BillPostalCd:          AddressMap["PostalCd"],
			BillCellPhone:         uuid.New(),
			BillEmail:             AddressMap["Email"],
			BillBirthdate:         uuid.New(),
			CCExpireMonth:         03,
			CCExpireYear:          2000 * randomDigitGen(2),
			CCId:                  randomDigitGen(8),
			ClientPaymentMethodId: uuid.New(),
			PaymentMethodNo:       one,
			Suffix:                uuid.New(),
			PaymentMethodType:     one,
		},
	}
}

func GetDefaultPromotionalPlanSet(promoCode string) []data.PromotionalPlanSet {
	return []data.PromotionalPlanSet{
		{
			PromoSetNo:   int64(PromoSetNo1),
			PromoSetName: PromoSetName1,
			PromoSetDesc: PromoSetDesc1,
			PromotionsForSet: []data.PromotionsForSet{
				{
					PromoCode:     promoCode,
					PromoCodeDesc: PromoCodeDesc1,
				},
			},
			ClientPromoSetId: ClientPromoSetId1,
		},
	}
}
func GetDefaultAllClientPlanDtl(promoCode string) []data.AllClientPlanDtl {
	return []data.AllClientPlanDtl{
		{
			PlanNo:                   int64(planNos[len(planNos)-1]),
			PlanName:                 planName[len(planName)-1],
			PlanDesc:                 PlanDesc1,
			SuppPlanInd:              0,
			BillingInd:               billingInd,
			DisplayInd:               displayInd,
			NewAcctStatus:            newAcctStatus,
			RolloverAcctStatus:       one,
			CurrencyCd:               Currency[0],
			ClientPlanId:             clientPlanId[len(clientPlanId)-1],
			ProrationInvoiceTimingCd: "I",
			RolloverPlanUomCd:        one,
			InitFreePeriodUomCd:      "1",
			InitialPlanStatusCd:      one,
			RolloverPlanStatusUomCd:  one,
			RolloverPlanStatusCd:     one,
			PlanServices:             GetDefaultPlanService(usageTypeCodes[0]),
			PromotionalPlanSets:      GetDefaultPromotionalPlanSet(promoCode),
			PlanSuppFields:           []data.PlanSuppField{},
		},
	}
}

func GetDefaultUsageTypes(idx int) data.UsageType {
	arr := []data.UsageType{
		{
			UsageTypeNo:   UsageTypeNo1,
			UsageTypeDesc: UsageTypeDesc1,
			UsageUnitType: "",
			UsageTypeName: UsageTypeName1,
			IsEditable:    false,
		},
		{
			UsageTypeNo:   UsageTypeNo2,
			UsageTypeDesc: UsageTypeDesc2,
			UsageUnitType: "",
			UsageTypeName: UsageTypeName2,
			IsEditable:    false,
		},
		{
			UsageTypeNo:   UsageTypeNo3,
			UsageTypeDesc: UsageTypeDesc3,
			UsageUnitType: "",
			UsageTypeName: UsageTypeName3,
			IsEditable:    false,
		},
	}
	return arr[idx]
}
func GetDefaultUsageUnitTypes(idx int) data.UsageUnitType {
	arr := []data.UsageUnitType{
		{
			UsageUnitTypeNo:   UsageUnitTypeNo1,
			UsageUnitTypeDesc: UsageUnitTypeDesc1,
		},
		{
			UsageUnitTypeNo:   UsageUnitTypeNo2,
			UsageUnitTypeDesc: UsageUnitTypeDesc2,
		},
		{
			UsageUnitTypeNo:   UsageUnitTypeNo3,
			UsageUnitTypeDesc: UsageUnitTypeDesc3,
		},
		{
			UsageUnitTypeNo:   UsageUnitTypeNo4,
			UsageUnitTypeDesc: UsageUnitTypeDesc4,
		},
		{
			UsageUnitTypeNo:   UsageUnitTypeNo5,
			UsageUnitTypeDesc: UsageUnitTypeDesc5,
		},
	}
	return arr[idx]
}
func GetDefaultUnappliedServiceCreditsDetail() []data.UnappliedServiceCreditsDetail {
	return []data.UnappliedServiceCreditsDetail{
		{
			CreateDate:             strcnv(time.Now().Format("00-00-0000")),
			CreateUser:             uuid.New(),
			InitialAmount:          float32(1000),
			AmountLeftToApply:      float32(1000),
			ReasonCd:               ReasonCd1,
			ReasonText:             ReasonText1,
			Comments:               Comments1,
			CurrencyCd:             Currency[0],
			ServiceNoToApply:       0,
			ServiceNameToApply:     ServiceNameToApply,
			ClientServiceIdToApply: "0",
			OutAcctNo:              acctNos[len(acctNos)-1],
			CreditId2:              creditId2,
		},
	}
}

func GetDefaultGetAcctResponse(clientAcctId string) response.GetAcctDetailsAllMResponse {
	return response.GetAcctDetailsAllMResponse{
		AriaResponse:            response.AriaResponse{ErrorCode: 0, ErrorMsg: "OK"},
		ClientAcctId:            clientAcctId,
		Userid:                  userIds[0],
		FirstName:               AddressMap["FirstName"],
		MiddleInitial:           "",
		LastName:                AddressMap["LastName"],
		CompanyName:             AddressMap["CompanyName"],
		Address1:                uuid.New(),
		Address2:                "",
		Address3:                "",
		City:                    AddressMap["City"],
		Locality:                "",
		StateProv:               AddressMap["StateProv"],
		CountryCd:               AddressMap["Country"],
		PostalCd:                AddressMap["PostalCd"],
		Phone:                   strcnv(randomDigitGen(8)),
		PhoneExt:                strcnv(randomDigitGen(3)),
		CellPhone:               strcnv(randomDigitGen(10)),
		Email:                   StringGen(8) + "@" + StringGen(5) + ".com",
		Birthdate:               "",
		StatusCd:                statusCd,
		NotifyMethod:            notifyMethod,
		SeqFuncGroupNo:          randomDigitGen(8),
		InvoiceApprovalRequired: invoiceApprovalRequired,
		FunctionalAcctGroup:     GetDefaultFunctionalAcctGroup(),
		AcctCurrency:            Currency[0],
		BillingGroupsInfo:       GetDefaultBillingGroupsInfo(),
		PaymentMethodsInfo:      []data.PaymentMethodsInfo{},
		MasterPlanCount:         one,
		MasterPlansInfo:         GetDefaultMasterPlansInfo(),
		AcctNo2:                 acctNos[len(acctNos)-1],
		ChiefAcctInfo:           GetDefaultChiefAcctInfo(),
	}
}

func GetDefaultPlanService(usageTypeCode string) []data.PlanService {
	return []data.PlanService{
		{
			ServiceNo:         int64(serviceNos[0]),
			ServiceDesc:       StringGen(25),
			IsRecurringInd:    0,
			IsUsageBasedInd:   one,
			UsageType:         int64(UsageTypeNo1),
			IsArrearsInd:      one,
			CoaId:             strcnv(randomDigitGen(8)),
			LedgerCode:        strcnv(randomDigitGen(8)),
			ClientCoaCode:     "1",
			DisplayInd:        one,
			TieredPricingRule: one,
			ClientServiceId:   clientServiceId[len(clientServiceId)-1],
			UsageTypeCd:       UsageTypeCd4,
			PlanServiceRates:  GetDefaultPlanServiceRate(),
			UsageTypeName:     UsageTypeName4,
			UsageTypeDesc:     UsageTypeDesc4,
			UsageTypeCode:     usageTypeCode,
			UsageUnitLabel:    "count",
		},
	}
}

func GetDefaultPlanServiceRate() []data.PlanServiceRate {
	return []data.PlanServiceRate{
		{
			RateSeqNo:            one,
			FromUnit:             1.0,
			RatePerUnit:          0,
			ClientRateScheduleId: StringGen(30),
		},
	}
}

func GetDefaultMasterPlansInfo() []data.MasterPlansInfo {
	masterPlanInstanceNo := randomDigitGen(7)
	return []data.MasterPlansInfo{
		{
			MasterPlanInstanceNo:       masterPlanInstanceNo,
			ClientMasterPlanInstanceId: strcnv(masterPlanInstanceNo),
			ClientMasterPlanId:         clientMasterPlanId,
			MasterPlanNo:               masterPlanNos[0],
		},
	}
}

func GetDefaultMasterPlanSummary() []data.MasterPlanSummary {
	return []data.MasterPlanSummary{{
		PlanInstanceNo:       planInstanceNos[len(planInstanceNos)-1],
		ClientPlanInstanceId: strcnv(planInstanceNos[len(planInstanceNos)-1]),
	},
	}
}

func GetDefaultAcctCredits(acctNo int64) []data.AllCredit {
	return []data.AllCredit{
		{
			OutAcctNo:               acctNo,
			OutMasterPlanInstanceNo: 0,
			OutClientMpInstanceId:   "",
			CreditNo:                randomDigitGen(8),
			CreatedBy:               "",
			CreatedDate:             "",
			Amount:                  1000,
			CreditType:              "",
			AppliedAmount:           0,
			UnappliedAmount:         0,
			ReasonCode:              0,
			ReasonText:              "",
			TransactionId:           0,
			VoidTransactionId:       0,
		},
	}
}

func GetDefaultCreditDetails(credit_no int64) response.GetCreditDetails {
	return response.GetCreditDetails{
		AriaResponse:             response.AriaResponse{},
		CreatedBy:                "",
		CreatedDate:              "",
		Amount:                   0,
		CreditType:               "",
		AppliedAmount:            0,
		UnappliedAmount:          0,
		ReasonCode:               0,
		ReasonText:               "",
		Comments:                 "",
		TransactionId:            0,
		VoidTransactionId:        0,
		CreditExpiryTypeInd:      "",
		CreditExpiryMonths:       0,
		CreditExpiryDate:         "",
		OutClientMpInstanceId:    "",
		AcctLocaleName:           "",
		OutAcctNo2:               0,
		OutMasterPlanInstanceNo2: 0,
		AcctLocaleNo2:            0,
		CreditExpiryPeriod:       0,
	}
}
