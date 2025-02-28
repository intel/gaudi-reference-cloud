// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Card, Button } from 'react-bootstrap'
import { BsArrowRightShort, BsCoin, BsWallet2 } from 'react-icons/bs'
import Spinner from '../../../utils/spinner/Spinner'
import { useNavigate } from 'react-router'
import { Link } from 'react-router-dom'

interface UsageWidgetProps {
  usageLoading: boolean
  creditsLoading: boolean
  totalAmount: number
  remainingCredits: number
  showCloudCredits: boolean
  usageError: boolean
  creditsError: boolean
}

const UsageWidget: React.FC<UsageWidgetProps> = ({
  usageLoading,
  creditsLoading,
  totalAmount,
  remainingCredits,
  showCloudCredits,
  usageError,
  creditsError
}): JSX.Element => {
  const spinner: JSX.Element = <Spinner />

  // navigate
  const navigate = useNavigate()

  const redirectTo = (route: string): void => {
    navigate(route)
  }

  const cardContent: JSX.Element = (
    <>
      {/* small view++ */}
      <div className="d-none d-sm-flex flex-column w-100 h-100 justify-content-center gap-s4">
        <div
          className="dashboard-item d-flex flex-row gap-s4 px-1"
          onClick={() => {
            redirectTo('/billing/usages')
          }}
        >
          <span className="my-1">
            <BsCoin />
          </span>
          <span className="my-auto">Current Month Usage</span>
          <span className={`my-auto ms-auto text-end ${usageError ? 'small' : 'fw-semibold'}`}>
            {usageError ? 'Data unavailable' : `${totalAmount} USD`}
          </span>
          <Link
            intc-id={'link-dashboard-usages'}
            data-wap_ref={'link-dashboard-usages'}
            aria-label={'Usages link'}
            className="btn btn-icon-simple btn-sm my-auto"
            to={'/billing/usages'}
          >
            <BsArrowRightShort />
          </Link>
        </div>
        {showCloudCredits ? (
          <div
            className="dashboard-item d-flex flex-row gap-s4 px-1"
            onClick={() => {
              redirectTo('/billing/credits')
            }}
          >
            <span className="my-1">
              <BsWallet2 />
            </span>
            <span className="my-auto">Remaining Credits</span>
            <span className={`my-auto ms-auto text-end ${usageError ? 'small' : 'fw-semibold'}`}>
              {creditsError ? 'Data unavailable' : `${remainingCredits} USD`}
            </span>
            <Link
              intc-id={'link-dashboard-credits'}
              data-wap_ref={'link-dashboard-credits'}
              aria-label={'Credits link'}
              className="btn btn-icon-simple btn-sm my-auto"
              to={'/billing/credits'}
            >
              <BsArrowRightShort />
            </Link>
          </div>
        ) : null}
      </div>

      {/* xs View */}
      <div className="d-none d-xs-flex d-sm-none w-100">
        <div className="d-flex flex-column gap-s4 w-100">
          <div
            className="dashboard-item d-flex flex-column"
            onClick={() => {
              redirectTo('/billing/usages')
            }}
          >
            <div className="d-flex flex-row gap-s4 px-1">
              <span className="my-1">
                <BsCoin />
              </span>
              <span className="my-auto">Current Month Usage</span>
            </div>
            <div className="d-flex flex-row gap-s4 px-1">
              <span className={`my-auto ms-auto text-end ${usageError ? 'small' : 'fw-semibold'}`}>
                {usageError ? 'Data unavailable' : `${totalAmount} USD`}
              </span>
              <Link
                intc-id={'link-dashboard-usages'}
                data-wap_ref={'link-dashboard-usages'}
                aria-label={'Usages link'}
                className="btn btn-icon-simple btn-sm my-auto"
                to={'/billing/usages'}
              >
                <BsArrowRightShort />
              </Link>
            </div>
          </div>
          {showCloudCredits ? (
            <div
              className="dashboard-item d-flex flex-column"
              onClick={() => {
                redirectTo('/billing/credits')
              }}
            >
              <div className="d-flex flex-row gap-s4 px-1">
                <span className="my-1">
                  <BsWallet2 />
                </span>
                <span className="my-auto">Remaining Credits</span>
              </div>
              <div className="d-flex flex-row gap-s4 px-1">
                <span className={`my-auto ms-auto text-end ${usageError ? 'small' : 'fw-semibold'}`}>
                  {creditsError ? 'Data unavailable' : `${remainingCredits} USD`}
                </span>
                <Link
                  intc-id={'link-dashboard-credits'}
                  data-wap_ref={'link-dashboard-credits'}
                  aria-label={'Credits link'}
                  className="btn btn-icon-simple btn-sm my-auto"
                  to={'/billing/credits'}
                >
                  <BsArrowRightShort />
                </Link>
              </div>
            </div>
          ) : null}
        </div>
      </div>

      <div className="d-flex ms-auto">
        <Button
          variant="outline-primary"
          intc-id="btn-redemm-coupon"
          data-wap_ref="btn-redemm-coupon"
          aria-label="btn-redemm-coupon"
          onClick={() => {
            redirectTo('/billing/credits/managecouponcode')
          }}
        >
          Redeem credit coupon
        </Button>
      </div>
    </>
  )

  return (
    <Card className="card-border-top w-100">
      <Card.Body className="h-100">
        <div className="d-flex justify-content-between w-100 align-items-center">
          <Card.Title>
            <h2 className="h6">Billing and Usage</h2>
          </Card.Title>
        </div>
        {usageLoading || creditsLoading ? spinner : cardContent}
      </Card.Body>
    </Card>
  )
}

export default UsageWidget
