// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"
)

// Set to true to have ReportUsage fail with a synthetic error
var FailReportUsage = false

// This is public for testing purposes
const MAX_USAGES_PER_QUERY = 2500

type CloudAccountToCost struct {
	Cost   float64
	Usages []int64
}

type Credit struct {
	CreditId        uint64
	Code            string
	RemainingAmount float64
}

func GetSortedCredits(ctx context.Context, session *sql.DB, cloudAccountId string) ([]Credit, error) {
	log := log.FromContext(ctx).WithName("DriverUtil.GetSortedCredits")
	log.Info("Executing ", "getSortedCredits for", cloudAccountId)
	defer log.Info("Returning", "getSortedCredits for", cloudAccountId)

	query := "SELECT id, coupon_code, remaining_amount " +
		"FROM cloud_credits " +
		"WHERE cloud_account_id=$1 AND ( expiry >= NOW() OR remaining_amount < 0 ) " +
		"ORDER BY created_at ASC "

	log.Info("get sorted credit", "query", query, "cloudAccount", cloudAccountId)
	var credits []Credit

	rows, err := session.QueryContext(ctx, query, cloudAccountId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		obj := Credit{}
		if err := rows.Scan(&obj.CreditId, &obj.Code, &obj.RemainingAmount); err != nil {
			return nil, err
		}
		log.Info("Credit", "creditId", obj.CreditId, "remainingAmount", obj.RemainingAmount)
		credits = append(credits, obj)
	}

	return credits, nil
}

func GetRemainingAmount(credits []Credit) float64 {
	totalAmt := float64(0)
	for _, obj := range credits {

		totalAmt += obj.RemainingAmount
	}
	return totalAmt
}

func ProcessCloudAccountCost(ctx context.Context, session *sql.DB, cloudAccountId string, cloudAccountToCost *CloudAccountToCost) error {
	log := log.FromContext(ctx).WithName("DriverUtil.ProcessCloudAccountCost")
	log.Info("Executing ProcessCloudAccountCost")
	defer log.Info("Returning from ProcessCloudAccountCost")

	creditList, err := GetSortedCredits(ctx, session, cloudAccountId)
	if err != nil {
		log.Error(err, "error getting credit from db")
		return err
	}

	totalRemainingAmount := GetRemainingAmount(creditList)
	log.Info("usage report", "cloudAccountId", cloudAccountId, "remainingAmount", totalRemainingAmount, "cost", cloudAccountToCost.Cost)

	tx, err := session.BeginTx(ctx, nil)
	if err != nil {
		log.Error(err, "error starting db transaction")
		return err
	}
	defer tx.Rollback()

	creditObj := Credit{}

	// Update remainingamount query
	query := "UPDATE cloud_credits " +
		"SET remaining_amount=$1, updated_at=NOW()::timestamp " +
		"WHERE cloud_account_id=$2 AND id=$3"

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		log.Error(err, "error in query")
		return err
	}
	defer stmt.Close()

	log.Info("Available credits", "creditlist", creditList)

	// Applying Greedy Approach for credit consumption wrt credit expiration
	for ii := 0; cloudAccountToCost.Cost > 0 && ii < len(creditList); ii++ {
		creditObj = creditList[ii]
		log.Info("Consuming", "creditId", creditObj.CreditId)
		var remainingAmount float64

		// if credits are installed during this iteration, then such credits will not have the
		// excess amount deducted when new credits are installed and hence can continue having a negative amount.
		// also we need to continue adding unreported cost
		if creditObj.RemainingAmount < 0 {
			absRemainingAmt := math.Abs(creditObj.RemainingAmount)
			if ii == (len(creditList) - 1) {
				remainingAmount = -(absRemainingAmt + cloudAccountToCost.Cost)
				cloudAccountToCost.Cost = 0
			} else {
				cloudAccountToCost.Cost += absRemainingAmt
				remainingAmount = 0
			}

		} else {
			if creditObj.RemainingAmount > cloudAccountToCost.Cost {
				remainingAmount = creditObj.RemainingAmount - cloudAccountToCost.Cost
				cloudAccountToCost.Cost = 0
			} else {
				if ii == (len(creditList) - 1) {
					remainingAmount = creditObj.RemainingAmount - cloudAccountToCost.Cost
					cloudAccountToCost.Cost = 0
				} else {
					cloudAccountToCost.Cost = cloudAccountToCost.Cost - creditObj.RemainingAmount
					remainingAmount = 0
				}
			}
		}

		_, err = stmt.ExecContext(ctx, remainingAmount, cloudAccountId, creditObj.CreditId)
		if err != nil {
			log.Error(err, "error updating db")
			return err
		}
	}

	// Remember the usages we processed for this cloudaccount
	for usages := cloudAccountToCost.Usages; len(usages) > 0; {
		numUsages := len(usages)
		if numUsages > MAX_USAGES_PER_QUERY {
			numUsages = MAX_USAGES_PER_QUERY
		}
		args, str := protodb.AddArrayArgValues([]any{}, usages[:numUsages],
			func(id int64) []any {
				return []any{id}
			})

		_, err = tx.ExecContext(ctx, "INSERT INTO credit_usage (usage_id) VALUES"+str, args...)
		if err != nil {
			return err
		}
		usages = usages[numUsages:]
	}

	if FailReportUsage {
		log.Info("synthetic error for testing")
		return errors.New("synthetic error for testing")
	}

	if err := tx.Commit(); err != nil {
		log.Error(err, "error committing db transaction")
		return err
	}
	return nil
}

func GetAlreadyBilledIds(ctx context.Context, session *sql.DB, startingUsageId int64) (map[int64]bool, error) {

	// Read all the usages in our table with ids >= the first id that
	// is being reported. This assumes that usages are reporting in
	// ascending id order
	query := "SELECT usage_id FROM credit_usage WHERE usage_id >= $1"
	rows, err := session.QueryContext(ctx, query, startingUsageId)
	if err != nil {
		return nil, err
	}
	ids := map[int64]bool{}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids[id] = true
	}

	return ids, nil
}

func ProcessCredit(ctx context.Context, session *sql.DB, creditList []Credit, cloudAccountId string) error {
	log := log.FromContext(ctx).WithName("DriverUtil.ProcessCredit")
	log.Info("Executing ProcessCredit")
	defer log.Info("Returning from ProcessCredit")

	lengthOfCreditList := len(creditList)
	if lengthOfCreditList > 1 {
		for i := lengthOfCreditList - 1; i > 0; i-- {
			if creditList[i-1].RemainingAmount < 0 && creditList[lengthOfCreditList-1].RemainingAmount > 0 {
				query := "UPDATE cloud_credits " +
					"SET remaining_amount=$1, updated_at=NOW()::timestamp " +
					"WHERE cloud_account_id=$2 AND id=$3"

				tx, err := session.BeginTx(ctx, nil)
				if err != nil {
					log.Error(err, "error starting db transaction")
					return err
				}
				defer tx.Rollback()
				stmt, err := tx.PrepareContext(ctx, query)
				if err != nil {
					log.Error(err, "error in preparing query for updating credits")
					return err
				}

				defer stmt.Close()

				remainingAmountForLastCredit := creditList[lengthOfCreditList-1].RemainingAmount - math.Abs(creditList[i-1].RemainingAmount)
				log.Info("credit remaining", "cloudAccountId", cloudAccountId, "remainingAmountForLastCredit", remainingAmountForLastCredit)
				_, err = stmt.ExecContext(ctx, 0, cloudAccountId, creditList[i-1].CreditId)
				if err != nil {
					log.Error(err, "error updating last credit with 0 when had negative remaining amount")
					return err
				}

				_, err = stmt.ExecContext(ctx, remainingAmountForLastCredit, cloudAccountId, creditList[lengthOfCreditList-1].CreditId)
				if err != nil {
					log.Error(err, "error updating added credit with remainining when had negative remaining amount")
					return err
				}
				if err := tx.Commit(); err != nil {
					log.Error(err, "error committing db transaction")
					return err
				}
			}
		}
	}
	return nil
}

func ParseRate(rate *pb.Rate) (float64, error) {
	if rate.GetRate() != "" {
		rateFloat, err := strconv.ParseFloat(rate.GetRate(), 64)
		if err != nil {
			return 0.0, fmt.Errorf("error parsing rates %v", err)
		}
		return rateFloat, err
	}
	return 0.0, fmt.Errorf("rate is empty")
}

func GetProductRateForAccountType(ctx context.Context, productId string, acctType pb.AccountType, productServiceClient *ProductClient) (*pb.Rate, error) {
	log := log.FromContext(ctx).WithName("DriverUtil.GetProductRateForAccountType")
	log.Info("get rate for product", "productId", productId)
	productFilter := pb.ProductFilter{
		AccountType: &acctType,
		Id:          &productId,
	}
	products, err := productServiceClient.GetProductCatalogProductsWithFilter(ctx, &productFilter)
	if err != nil {
		return nil, err
	}
	if len(products) > 1 || len(products) == 0 {
		productErr := fmt.Errorf("error in getting rate for product id %v", products)
		log.Error(productErr, "error getting rates", "products", products)
		return nil, productErr
	}
	rates := products[0].GetRates()
	if len(rates) > 1 {
		rateErr := fmt.Errorf("error in getting rate due to multiple rate for product id")
		log.Error(rateErr, "error getting rates", "productId", productId, "rates", rates)
		//TODO: fix product catalog mock data
		return rates[0], nil
	}
	log.Info("return rate ", "rate", rates[0])
	return rates[0], nil
}
