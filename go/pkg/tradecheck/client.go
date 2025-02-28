// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tradecheck

import (
	"context"
	"fmt"
)

type GTSclient struct{}

type config struct{}

func NewClient(cfg config) (*GTSclient, error) {
	return &GTSclient{}, nil
}

func CreateConfig(usernameFile string, passwordFile string, tokenURL string, createProductURL string, createOrderURL string, screenBusinessPartnerURL string) (config, error) {
	return config{}, nil
}

func (d *GTSclient) CreateProduct(ctx context.Context, product Product) error {
	fmt.Println("Default CreateProduct called")
	return nil
}

func (d *GTSclient) CreateOrder(ctx context.Context, order Order) (CreateOrderResponse, error) {
	fmt.Println("Default CreateOrder called")
	return CreateOrderResponse{}, nil
}

func (d *GTSclient) IsOrderValid(ctx context.Context, productId, email, personId, countryCode string) (bool, error) {
	fmt.Println("Default IsOrderValid called")
	return true, nil
}

func (d *GTSclient) ScreenBusinessPartner(ctx context.Context, screenRequest ScreenRequest) (ScreenResponse, error) {
	fmt.Println("Default ScreenBusinessPartner called")
	return ScreenResponse{}, nil
}
