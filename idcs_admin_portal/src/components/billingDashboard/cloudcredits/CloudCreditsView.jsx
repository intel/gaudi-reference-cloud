import React, { useEffect, useState } from 'react'

import { Button, Col, Row } from 'react-bootstrap'
import { PencilFill } from 'react-bootstrap-icons'
import GridPagination from '../../../utility/gridPagination/gridPagination'

const moment = require('moment')

const CloudCreditsView = (props) => {
  const [cloudcreditData, setCloudcreditData] = useState([])
  const [filterValue] = useState('')
  const [data, setData] = useState([])
  const [selectedRecord, setSelectedRecord] = useState([])

  const columns = [
    { columnName: 'Account' },
    { columnName: 'Account type' },
    { columnName: 'Cloud Credit code' },
    { columnName: 'Redeemed Date' },
    { columnName: 'Status' },
    { columnName: 'Credit amount' },
    { columnName: 'Credit used' },
    { columnName: 'Remaining credit' }
  ]

  useEffect(() => {
    const cloudcredits = props.getCloudCredits()
    setCloudcreditData(cloudcredits)
    setData(buildTable(filterValue, cloudcredits.data))
    const filterSingleRecord = cloudcredits.data.filter((item) =>
      item.code.includes(props.code)
    )

    setSelectedRecord(filterSingleRecord)
  }, [props])

  // function filterData(searchKey = "") {
  //   setFilterValue(searchKey);
  //   setData(buildTable(searchKey));
  // }

  function buildTable (searchKey, tempCloudCreditData = cloudcreditData.data) {
    const table = tempCloudCreditData.map((row) => {
      const rowData = {}
      if (
        row.code.toString().toLowerCase().indexOf(searchKey.toLowerCase()) >
          -1 ||
        moment(row.redeemed_date)
          .format('MM/DD/YYYY hh:mm:ss')
          .toString()
          .toLowerCase()
          .indexOf(searchKey.toLowerCase()) > -1 ||
        row.credit_amount.toLowerCase().indexOf(searchKey.toLowerCase()) > -1 ||
        row.times_redeemed.toLowerCase().indexOf(searchKey.toLowerCase()) >
          -1 ||
        row.credits_awarded.toLowerCase().indexOf(searchKey.toLowerCase()) >
          -1 ||
        searchKey === ''
      ) {
        rowData.account = row.account.toString()
        rowData.account_type = row.account_type.toString()
        rowData.code = row.code.toString()
        rowData.redeemed_date = row.redeemed_date.toString()
        rowData.status = row.status.toString()
        rowData.credit_amount = row.credit_amount
        rowData.credit_used = row.credit_used
        rowData.remaining_credit = row.remaining_credit
      }
      return rowData
    })
    return table
  }

  return (
    <>
      <Row className="ps-2 mb-2 col-12">
        <span className="col-6 pt-4 pl-2">
          <h2 className="pt-2">&nbsp;View Cloud Credits: {props.code}</h2>
        </span>
        <span className="col-6 text-end pt-4">
          <Button
            variant="link"
            onClick={() => props.whichPageToOpen('cloudcredits')}
            className="btn-sm"
          >
            Back to list
          </Button>
        </span>
      </Row>
      <Row className="ps-2 mb-2 col-12">
        <span className="col-6 pt-4 pl-2">
          <h3 className="pt-2">&nbsp;Cloud credit details</h3>
        </span>

        <span className="col-6 text-end pt-4">
          <Button
            variant="outline-primary"
            onClick={() => props.whichPageToOpen('edit', props.code)}
            className="btn-sm"
          >
            <PencilFill />
            &nbsp;&nbsp;&nbsp;Edit
          </Button>
          &nbsp;&nbsp;&nbsp;&nbsp;
          <Button
            variant="outline-primary"
            onClick={() => props.whichPageToOpen('deactivate')}
            className="btn-sm"
          >
            Deactivate code
          </Button>
        </span>
      </Row>
      <div className="col-12 ps-3 pb-4">
        {/* <Row>
          <b>Credit name</b>
          <p>{selectedRecord[0]?.name}</p>
        </Row> */}
        <Row>
          <b>Credit code</b>
          <p>{selectedRecord[0]?.code}</p>
        </Row>
        <Row>
          <b>Amount</b>
          <p>{selectedRecord[0]?.credit_amount}</p>
        </Row>
        <Row>
          <b>Credit description</b>
          <p>{selectedRecord[0]?.description}</p>
        </Row>
        <Row>
          <b>Audience</b>
          <p>
            {selectedRecord[0]?.audience === 'wildcard'
              ? props.form.audience.wildcard
              : selectedRecord[0]?.audience}
          </p>
        </Row>
        <Row>
          <b>Expires on</b>
          <p>
            {moment(selectedRecord[0]?.expiry_date)
              .format('MM/DD/YYYY hh:mm:ss')
              .toString()}
          </p>
        </Row>
        <Row>
          <b>Status</b>
          <p>{selectedRecord[0]?.status}</p>
        </Row>
      </div>
      {/* {Commented filter field } */}
      {/* <Row className="col-12 ps-4 pt-3 mb-5">
        <div className="col-4 ps-0">
          <label className="col-12 ps-1">Filter credits</label>
          <div className="input-group has-validation col-12 ps-1">
            <div className="input-group-text  border-right-0 bg-white rounded-0">
              <Icon.Search color="black" className="" />
            </div>
            <input
              type="text"
              className="form-control border-left-0"
              required=""
              onChange={(e) => filterData(e.target.value)}
            />
          </div>
        </div>
      </Row> */}
      <h3 className="col pt-2 pb-4">&nbsp;&nbsp;&nbsp;Cloud credit Usage</h3>
      <Row className="pt-0 ps-4 pe-4 mb-4">
        <Col className="ps-2">
          <GridPagination
            data={data}
            columns={columns}
            loading={props.loading}
          />
        </Col>
      </Row>
    </>
  )
}

export default CloudCreditsView
