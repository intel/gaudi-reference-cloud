// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { BsCopy } from 'react-icons/bs'
import Button from 'react-bootstrap/Button'
import Wrapper from '../Wrapper'
import CustomAlerts from '../customAlerts/CustomAlerts'
import LabelValuePair from '../labelValuePair/LabelValuePair'

const TapContent = (props) => {
  // props
  const infoToDisplay = props.infoToDisplay
  const title = infoToDisplay.tapTitle
  const hiddenTitle = infoToDisplay.hiddenTitle ?? 'Selected Tab'
  const fields = infoToDisplay.fields ? infoToDisplay.fields : []
  const tapConfigType = infoToDisplay.tapConfig.type
  const subTitle = infoToDisplay.tapConfig.subTitle
  const headers = infoToDisplay.tapConfig.headers
  const customContent = infoToDisplay.customContent
  const columnCount = parseInt(infoToDisplay.tapConfig.columnCount)
  const rowsCount = Math.ceil(fields.length / columnCount)
  const columnSize = Math.round(12 / columnCount)

  // functions
  function getData(rowIndex, colIndex) {
    let result = null

    const fieldIndex = colIndex + rowIndex * columnCount

    if (fields[fieldIndex]) {
      const field = fields[fieldIndex]
      const value = field.value

      result = (
        <LabelValuePair className={`col-md-${columnSize}`} key={colIndex} label={field.label}>
          {value}
          {field.action
            ? field.action.map((item, index) => (
                <Button
                  variant="icon-simple"
                  aria-label={`Copy ${field.label ? field.label.replace(':', '').trimEnd() : 'text'}`}
                  onClick={() => item.func(value)}
                  key={index}
                >
                  <BsCopy className="text-primary" />
                </Button>
              ))
            : null}
        </LabelValuePair>
      )
    }

    return result
  }

  function buildInformation() {
    const rows = []

    for (let index = 0; index < rowsCount; index++) {
      rows.push(index)
    }

    const columns = []
    for (let index = 0; index < columnCount; index++) {
      columns.push(index)
    }

    return (
      <>
        {rows.map((_, rowIndex) => (
          <div className="row" key={rowIndex}>
            {columns.map((colIndex) => getData(rowIndex, colIndex))}
          </div>
        ))}
      </>
    )
  }

  function getTableData(value) {
    let result = <p>{value}</p>

    if (Array.isArray(value)) {
      result = value.map((item, index) => <p key={index}>{item}</p>)
    }

    return result
  }

  function buildTable() {
    const columnSize = 12 / columnCount

    let subTitleContent = null

    if (subTitle) {
      subTitleContent = (
        <div className="row ms-0">
          <CustomAlerts showAlert={true} alertType="secondary" message={subTitle} onCloseAlert={null} showIcon={true} />
        </div>
      )
    }

    const result = (
      <Wrapper>
        {subTitleContent}
        <div className="col-6 row ms-0 border border-light">
          {headers.map((header, index) => (
            <div className={`col-${columnSize} border-bottom`} key={index}>
              <p className="fw-semibold">{header}</p>
            </div>
          ))}
        </div>
        <div className="col-6 row ms-0 border border-light">
          {fields.map((field, index) => (
            <div className={`col-${columnSize} border-bottom`} key={index}>
              {getTableData(field.value)}
            </div>
          ))}
        </div>
      </Wrapper>
    )

    return result
  }

  // Varaibles
  let tapDisplay = null

  switch (tapConfigType) {
    case 'columns':
      tapDisplay = buildInformation()
      break
    case 'table':
      tapDisplay = buildTable()
      break
    case 'custom':
      tapDisplay = customContent
      break
    default:
      break
  }

  return (
    <>
      {!title && (
        <h3 aria-label={hiddenTitle} className="m-ns6 p-0" style={{ fontSize: 0 }}>
          {hiddenTitle}
        </h3>
      )}
      <div className="section px-s0 px-sm-s5">
        {title && <h3>{title}</h3>}
        {tapDisplay}
      </div>
    </>
  )
}

export default TapContent
