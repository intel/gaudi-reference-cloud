import React, { useEffect, useState } from 'react'

import { Alert, Button, Col, Row } from 'react-bootstrap'
import GridPagination from '../../../utility/gridPagination/gridPagination'
import * as Icon from 'react-bootstrap-icons'
import EmptyView from '../../../utility/emptyView/EmptyView'

const moment = require('moment')

const CloudCreditsView = (props) => {
  const [cloudcreditData, setCloudcreditData] = useState([])
  const [filterValue, setFilterValue] = useState('')
  const [data, setData] = useState([])
  const [createShow, setCreateShow] = useState(true)

  const columns = [
    { name: 'creditCode', columnName: 'Cloud Credit code' },
    { name: 'description', columnName: 'Description' },
    { name: 'audience', columnName: 'Audience' },
    { name: 'billing_expiry_dateperiod', columnName: 'Expiration Date' },
    { name: 'status', columnName: 'Status' },
    { name: 'creditUnitAmount', columnName: 'Credit amount', className: 'text-end' },
    { name: 'times_redeemed', columnName: 'Times redeemed' },
    { name: 'credits_awarded', columnName: 'Credits awarded' }
  ]

  useEffect(() => {
    const cloudcredits = props.cloudCredits
    setCloudcreditData(cloudcredits)
    setData(buildTable(filterValue, cloudcredits))
  }, [props])

  function filterData (searchKey = '') {
    setFilterValue(searchKey)
    setData(buildTable(searchKey))
  }

  function formatAmount (amount, currency = 'USD') {
    return amount.toLocaleString('en-US', { style: 'currency', currency })
  }

  function buildTable (searchKey, tempCloudCreditData = cloudcreditData) {
    const table = tempCloudCreditData.map((row) => {
      const rowData = {}
      if (
        row.creditCode
          .toString()
          .toLowerCase()
          .indexOf(searchKey.toLowerCase()) > -1 ||
        moment(row.expirationTime)
          .format('MM/DD/YYYY hh:mm:ss')
          .toString()
          .toLowerCase()
          .indexOf(searchKey.toLowerCase()) > -1 ||
        row.creditUnitAmount.toString().indexOf(searchKey.toLowerCase()) > -1 ||
        // row.times_redeemed.toLowerCase().indexOf(searchKey.toLowerCase()) >
        //   -1 ||
        // row.credits_awarded.toLowerCase().indexOf(searchKey.toLowerCase()) >
        //   -1 ||
        searchKey === ''
      ) {
        rowData.creditCode = (
          <Button
            variant="link"
            onClick={() => props.whichPageToOpen(row.creditCode.toString())}
          >
            {row.creditCode.toString()}
          </Button>
        )
        rowData.description = row.description.toString()
        rowData.audience = 'TBD-API'
        // row.audience.toString() === "wildcard"
        //   ? props.form.audience.wildcard
        //   : row.audience.toString();
        rowData.billing_expiry_dateperiod = row.expirationTime ? moment(row.expirationTime).format('MM/DD/YYYY hh:mm:ss').toString() : 'TBD-API'
        rowData.status = 'TBD-API'
        rowData.creditUnitAmount = formatAmount(row.creditUnitAmount, row.currency)
        rowData.times_redeemed = 'TBD-API'
        rowData.credits_awarded = 'TBD-API'
      }
      return rowData
    })

    return table
  }

  return (
    <>
      <Row className="ps-2 mb-2 col-12">
        <span className="col-6 pt-4 pl-2">
          <h2 className="pt-2">&nbsp;Manage Cloud Credits</h2>
        </span>

        <span className="col-6 text-end pt-4">
          <Button
            variant="primary"
            onClick={() => props.whichPageToOpen('create')}
            className="btn-sm"
            intc-id="button-cloudcredits-create"
          >
            Create Cloud credit code
          </Button>
        </span>
      </Row>

      <div className="col-12 ps-3 pt-0">
        <p>As of TBD-API</p>
      </div>
      {props.createdCCMessage && createShow && (
        <Alert
          variant="success"
          className="col-4 center"
          onClose={() => setCreateShow(false)}
          dismissible
          intc-id="alert-cloudcredits-create"
        >
          Cloud credit &apos;{props.createdCCMessage}&apos; was created
        </Alert>
      )}

      { data.length > 0
        ? (
      <>
      <Row className="col-12 ps-4 pt-3 mb-5">
        <div className="col-4 ps-0">
          <label className="col-12 ps-1">Filter credits</label>
          <div className="input-group has-validation col-12 ps-1">
            <div className="input-group-text  border-right-0 bg-white rounded-0">
              <Icon.Search color="black" className="" />
            </div>
            <input
              type="text"
              className="form-control border-left-0"
              intc-id="input-cloudcredits-filter"
              required=""
              onChange={(e) => filterData(e.target.value)}
            />
          </div>
        </div>
      </Row>

      <Row className="pt-0 ps-4 pe-4" intc-id="card-cloudcredits-list">
        <Col className="ps-2">
          <GridPagination
            data={data}
            columns={columns}
            loading={false}
          />
        </Col>
      </Row>
      </>
          )
        : props.loading
          ? (
        <div className="row col-12">
          <div className="spinner-border text-primary center"></div>
        </div>
            )
          : (
        <div className="ps-3 pe-3" intc-id="card-cloudcredits-empty">
          <div className="row justify-content-center mb-5 ps-3">
            <div className="ps-0">
              <EmptyView title={'Items not found'} subTitle={null} action={null} />
            </div>
          </div>
        </div>
            )}
    </>
  )
}

export default CloudCreditsView
