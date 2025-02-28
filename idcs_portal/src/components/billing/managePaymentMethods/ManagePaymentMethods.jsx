// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import Button from 'react-bootstrap/Button'
import ButtonGroup from 'react-bootstrap/ButtonGroup'
import Card from 'react-bootstrap/Card'
import { CreditCardsTypes } from '../../../utils/Enums'
import Spinner from '../../../utils/spinner/Spinner'

const ManagePaymentMethods = (props) => {
  const cloudCredits = props.cloudCredits
  const cardCredits = props.cardCredits
  const title = props.title
  const helperTitle = props.helperTitle
  const loading = props.loading
  const creditMethods = []
  const cloudMethods = []
  cloudMethods.push(cloudCredits)
  creditMethods.push(cardCredits)

  return (
    <>
      <div className="section">
        <h2 intc-id="managePaymentMethodsTitle">{title}</h2>
        <span>{helperTitle}</span>
        <div className="d-flex flex-xs-column flex-md-row gap-s8">
          <>
            {creditMethods.map((method, index) => (
              <Card key={index}>
                <Card.Header as="h3">{method.title}</Card.Header>
                <Card.Body>
                  {loading ? (
                    <Spinner />
                  ) : method.cardNbr ? (
                    <>
                      <h4>{method.subTitle}</h4>
                      <div className="d-flex flex-row gap-s6 text-nowrap">
                        <span className="fw-semibold"> {method.cardHolderName}</span>
                        <span className="text-nowrap">Ends in {method.cardNbr}</span>
                        <span className="text-nowrap">Exp {method.expDt}</span>
                        <div
                          className={`${
                            method.brandtype === CreditCardsTypes.MasterCardCreditCard
                              ? 'creditCard-mastercard'
                              : method.brandtype === CreditCardsTypes.VisaCreditCard
                                ? 'creditCard-visa'
                                : method.brandtype === CreditCardsTypes.DiscoverCreditCard
                                  ? 'creditCard-discover'
                                  : method.brandtype === CreditCardsTypes.AmexCreditCard
                                    ? 'creditCard-amex'
                                    : 'creditCard-mastercard'
                          } creditCard-all-span`}
                        >
                          {' '}
                          <span></span>{' '}
                        </div>
                      </div>
                      <ButtonGroup className="mt-auto pt-s6">
                        {method.actions.map((action, index) => (
                          <Button
                            key={index}
                            variant="outline-primary"
                            intc-id="btn-managepayment-changecard"
                            data-wap_ref="btn-managepayment-changecard"
                            aria-label={action.buttonLabel}
                            onClick={() => action.function(method.cardNbr)}
                          >
                            {method.cardNbr ? 'Change card' : action.buttonLabel}
                          </Button>
                        ))}
                      </ButtonGroup>
                      <span className="valid-feedback">*To delete your credit card, contact the support team.</span>
                    </>
                  ) : (
                    <>
                      <h4>{method.helperMessage}</h4>
                      <ButtonGroup className="mt-auto pt-s6">
                        {method.actions.map((action, index) => (
                          <Button
                            key={index}
                            variant="outline-primary"
                            intc-id="btn-managepayment-addcard"
                            aria-label={action.buttonLabel}
                            data-wap_ref="btn-managepayment-addcard"
                            onClick={() => action.function(method.cardNbr)}
                          >
                            {method.cardNbr ? 'Change card' : action.buttonLabel}
                          </Button>
                        ))}
                      </ButtonGroup>
                    </>
                  )}
                </Card.Body>
              </Card>
            ))}
            {cloudMethods.map((method, index) => (
              <Card key={index}>
                <Card.Header as="h3">{method.title}</Card.Header>
                <Card.Body>
                  {loading ? (
                    <Spinner />
                  ) : method.balance ? (
                    <>
                      <h4>{method.subTitle}</h4>
                      <div className="d-flex flex-row gap-s6 text-nowrap">
                        <span className="fw-semibold">{method.balance}</span>
                        <span>Expiration date {method.expDt}</span>
                      </div>
                      <ButtonGroup className="mt-auto pt-s6">
                        {method.actions.map((action, index) => (
                          <Button
                            key={index}
                            variant="outline-primary"
                            aria-label={action.buttonLabel}
                            intc-id="btn-managepayment-addcredits"
                            data-wap_ref="btn-managepayment-addcredits"
                            onClick={() => action.function()}
                          >
                            {action.buttonLabel}
                          </Button>
                        ))}
                      </ButtonGroup>
                    </>
                  ) : (
                    <>
                      <h4>{method.helperMessage}</h4>
                      <ButtonGroup>
                        {method.actions.map((action, index) => (
                          <Button
                            variant="outline-primary"
                            key={index}
                            intc-id="btn-managepayment-addcredits"
                            data-wap_ref="btn-managepayment-addcredits"
                            aria-label={action.buttonLabel}
                            onClick={() => action.function()}
                          >
                            {action.buttonLabel}
                          </Button>
                        ))}
                      </ButtonGroup>
                    </>
                  )}
                </Card.Body>
              </Card>
            ))}
          </>
        </div>
      </div>
    </>
  )
}

export default ManagePaymentMethods
