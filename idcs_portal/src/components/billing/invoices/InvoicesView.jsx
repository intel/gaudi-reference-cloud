// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import Button from 'react-bootstrap/Button'
import GridPagination from '../../../utils/gridPagination/gridPagination'
import CustomInput from '../../../utils/customInput/CustomInput'
import { Link } from 'react-router-dom'
import { Form } from 'react-bootstrap'
import CustomInputLabel from '../../../utils/customInput/partials/CustomInputLabel'
import SearchBox from '../../../utils/searchBox/SearchBox'
import Spinner from '../../../utils/spinner/Spinner'

const InvoicesView = (props) => {
  // props
  const state = props.state
  const form = state.form
  const columns = props.columns
  const myInvoices = props.filteredInvoices
  const currentDateTime = props.currentDateTime
  const hasInvoices = props.hasInvoices

  // props function
  const onChangeInput = props.onChangeInput

  // variables
  const configInput = { ...form.filterInvoices }
  const configStartDate = { ...form.filterStartDate }
  const configEndDate = { ...form.filterEndDate }

  const emptyGridByFilter = {
    title: 'No invoices found',
    subTitle: 'The applied filter criteria did not match any items.',
    action: {
      type: 'function',
      href: () => refreshFilters(),
      label: 'Clear filters'
    }
  }

  function refreshFilters() {
    onChangeInput(null, 'clearFilters')
  }

  return (
    <>
      {props.loading ? (
        <Spinner />
      ) : (
        <>
          {!hasInvoices ? (
            <div className="section text-center align-items-center">
              <h1>No invoices found</h1>
              <p>Your account currently has no invoice to report for this month.</p>
            </div>
          ) : (
            <>
              <div className="section">
                <h2 intc-id="invoicesTitle">Invoices</h2>
                <span>As of {currentDateTime} </span>
              </div>
              <div className="section">
                <h3 className="h3">Invoice History</h3>
                <span>
                  For current month data, view &nbsp;
                  <Link
                    intc-id="usagesButton"
                    aria-label="Go to current month usage page"
                    className="link"
                    to="/billing/usages"
                  >
                    Current month usage
                  </Link>
                  .
                </span>
                <div className="d-flex flex-wrap gap-s6">
                  <Form.Group className="d-flex-customInput w-auto">
                    <CustomInputLabel label={configInput.label} />
                    <SearchBox
                      intc-id="FilterProductsInput"
                      aria-label={configInput.label}
                      value={configInput.value}
                      placeholder="Type to filter invoices..."
                      onChange={(event) => onChangeInput(event, 'filterInvoices')}
                    />
                  </Form.Group>
                  <div className="d-flex flex-row gap-s6">
                    <CustomInput
                      key="filterStartDate"
                      type={configStartDate.type}
                      fieldSize={configStartDate.fieldSize}
                      isRequired={configStartDate.validationRules.isRequired}
                      label={
                        configStartDate.validationRules.isRequired
                          ? configStartDate.label + ' *'
                          : configStartDate.label
                      }
                      value={configStartDate.value}
                      onChanged={(event) => onChangeInput(event, 'filterStartDate')}
                      isValid={configStartDate.isValid}
                      isTouched={configStartDate.isTouched}
                      isReadOnly={configStartDate.isReadOnly}
                      showMonthDropdown={configStartDate.showMonthDropdown}
                      showYearDropdown={configStartDate.showYearDropdown}
                      dateFormat={configStartDate.dateFormat}
                      startDate={
                        Object.prototype.hasOwnProperty.call(configStartDate, 'startDate')
                          ? form[configStartDate.startDate].value
                            ? form[configStartDate.startDate].value
                            : false
                          : false
                      }
                      endDate={
                        Object.prototype.hasOwnProperty.call(configStartDate, 'endDate')
                          ? form[configStartDate.endDate].value
                            ? form[configStartDate.endDate].value
                            : false
                          : false
                      }
                      maxDate={
                        Object.prototype.hasOwnProperty.call(configStartDate, 'maxDate')
                          ? form[configStartDate.maxDate].value
                            ? form[configStartDate.maxDate].value
                            : false
                          : false
                      }
                      selectsStart={configStartDate.selectsStart}
                    />
                    <CustomInput
                      key="filterEndDate"
                      type={configEndDate.type}
                      fieldSize={configEndDate.fieldSize}
                      isRequired={configEndDate.validationRules.isRequired}
                      label={
                        configEndDate.validationRules.isRequired ? configEndDate.label + ' *' : configEndDate.label
                      }
                      value={configEndDate.value}
                      onChanged={(event) => onChangeInput(event, 'filterEndDate')}
                      isValid={configEndDate.isValid}
                      isTouched={configEndDate.isTouched}
                      isReadOnly={configEndDate.isReadOnly}
                      showMonthDropdown={configEndDate.showMonthDropdown}
                      showYearDropdown={configEndDate.showYearDropdown}
                      dateFormat={configEndDate.dateFormat}
                      startDate={
                        Object.prototype.hasOwnProperty.call(configEndDate, 'startDate')
                          ? form[configEndDate.startDate].value
                            ? form[configEndDate.startDate].value
                            : false
                          : false
                      }
                      endDate={
                        Object.prototype.hasOwnProperty.call(configEndDate, 'endDate')
                          ? form[configEndDate.endDate].value
                            ? form[configEndDate.endDate].value
                            : false
                          : false
                      }
                      minDate={
                        Object.prototype.hasOwnProperty.call(configEndDate, 'minDate')
                          ? form[configEndDate.minDate].value
                            ? form[configEndDate.minDate].value
                            : false
                          : false
                      }
                      selectsEnd={configEndDate.selectsEnd}
                    />
                  </div>
                  <div className="d-flex mt-auto">
                    <Button
                      intc-id="clearFilterButton"
                      data-wap_ref="clearFilterButton"
                      aria-label="Clear filters"
                      variant="outline-primary"
                      onClick={() => refreshFilters()}
                    >
                      Clear filters
                    </Button>
                  </div>
                </div>
              </div>
              <div className="section">
                <GridPagination data={myInvoices} emptyGrid={emptyGridByFilter} columns={columns} loading={false} />
              </div>
            </>
          )}
        </>
      )}
    </>
  )
}

export default InvoicesView
