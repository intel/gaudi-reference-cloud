import React from 'react'
import CustomInput from '../../utility/customInput/CustomInput'
import Button from 'react-bootstrap/Button'
import { ButtonGroup } from 'react-bootstrap'
import OnSubmitModal from '../../utility/modals/onSubmitModal/OnSubmitModal'

const NodeEdit = (props) => {
  const nodeDetails = props.nodeDetails
  const state = props.state
  const title = state.title
  const description = state.description
  const form = state.form
  const navigationTop = state.navigationTop
  const navigationBottom = state.navigationBottom
  const isValidForm = state.isValidForm
  const onChangeInput = props.onChangeInput
  const onSubmit = props.onSubmit
  const onChangeDropdownMultiple = props.onChangeDropdownMultiple
  const onSelectAll = props.onSelectAll
  const showModal = props.showModal
  const apiCallsCompleted = props.apiCallsCompleted

  const instanceTypeGroup = []

  const poolGroup = []

  for (const key in form) {
    const formItem = {
      ...form[key]
    }

    if (formItem.sectionGroup === 'instanceType') {
      if (key === 'instanceTypes') {
        formItem.selectAllButton = {
          label: 'Select/Deselect All',
          buttonFunction: () => onSelectAll('instanceTypes')
        }
      }
      instanceTypeGroup.push({
        id: key,
        configInput: formItem
      })
    } else if (formItem.sectionGroup === 'pool') {
      if (key === 'pools') {
        formItem.selectAllButton = {
          label: 'Select/Deselect All',
          buttonFunction: () => onSelectAll('pools')
        }
      }
      poolGroup.push({
        id: key,
        configInput: formItem
      })
    }
  }

  function buildCustomInput(element) {
    return (
      <CustomInput
        key={element.id}
        type={element.configInput.type}
        fieldSize={element.configInput.fieldSize}
        placeholder={element.configInput.placeholder}
        isRequired={element.configInput.validationRules.isRequired}
        label={
          element.configInput.validationRules.isRequired ? element.configInput.label + ' *' : element.configInput.label
        }
        value={element.configInput.value}
        onChanged={(event) => onChangeInput(event, element.id)}
        isValid={element.configInput.isValid}
        isTouched={element.configInput.isTouched}
        helperMessage={element.configInput.helperMessage}
        isReadOnly={element.configInput.isReadOnly}
        options={element.configInput.options}
        validationMessage={element.configInput.validationMessage}
        readOnly={element.configInput.readOnly}
        refreshButton={element.configInput.refreshButton}
        isMultiple={element.configInput.isMultiple}
        onChangeSelectValue={element.configInput.onChangeSelectValue}
        extraButton={element.configInput.extraButton}
        emptyOptionsMessage={element.configInput.emptyOptionsMessage}
        borderlessDropdownMultiple={element.configInput.borderlessDropdownMultiple}
        onChangeDropdownMultiple={(values) => onChangeDropdownMultiple(values, element.id)}
        selectAllButton={element.configInput.selectAllButton}
        customContent={element.configInput.customContent}
        hideContent={false}
        maxWidth={element.configInput.maxWidth}
      />
    )
  }

  return (
    <>
      <div className="section">
        {navigationTop.map((item, index) => (
          <div className="m-lg-0" key={index}>
            {
              <Button intc-id={`navigationTop${item.label}`} variant={item.buttonVariant} onClick={item.function}>
                {item.label}
              </Button>
            }
          </div>
        ))}
      </div>
      <OnSubmitModal showModal={showModal} message="Working on your request" />

      <div className="section">
        <h1 className="h2">{title}</h1>
        <h2 className="h4">{description}</h2>
      </div>
      {nodeDetails && (
        <div className="section">
          <div className="row mb-2">
            <div className="col-12 col-md-2 fw-bold">Node Name:</div>
            <div className="col-12 col-md-4">{nodeDetails.nodeName}</div>
          </div>
          <div className="row mb-2">
            <div className="col-12 col-md-2 fw-bold">Region:</div>
            <div className="col-12 col-md-4">{nodeDetails.region}</div>
          </div>
          <div className="row mb-2">
            <div className="col-12 col-md-2 fw-bold">Availability zone:</div>
            <div className="col-12 col-md-4">{nodeDetails.availabilityZone}</div>
          </div>
          <div className="row mb-2">
            <div className="col-12 col-md-2 fw-bold">Resources Used:</div>
            <div className="col-12 col-md-4">{nodeDetails.percentageResourcesUsed}%</div>
          </div>
          <div className="row mb-2">
            <div className="col-12 col-md-2 fw-bold">Instance Types:</div>
            <div className="col-12 col-md-10">{nodeDetails?.instanceTypes.join(', ')}</div>
          </div>
          <div className="row mb-2">
            <div className="col-12 col-md-2 fw-bold">Compute Node Pools:</div>
            <div className="col-12 col-md-10">{nodeDetails?.poolIds.join(', ')}</div>
          </div>
          <hr />
        </div>
      )}

      <div className="section">
        <h2 className="h4">Override the Node values</h2>
        <div className="row">
          <div className="col-12 fw-bold">Instance Types</div>
          <div className="col-12 col-md-12">
            {instanceTypeGroup.map((element, index) => (
              <div className="col-12" key={index}>
                {buildCustomInput(element)}
              </div>
            ))}
          </div>
        </div>
        <div className="row">
          <div className="col-12 fw-bold">Compute Node Pools</div>
          <div className="col-12 col-md-12">
            {poolGroup.map((element, index) => (
              <div className="col-12" key={index}>
                {buildCustomInput(element)}
              </div>
            ))}
          </div>
        </div>

        <ButtonGroup>
          {navigationBottom.map((item, index) => (
            <Button
              key={index}
              intc-id={`navigationTop${item.label}`}
              disabled={item.label === 'Save' ? !isValidForm || !apiCallsCompleted : false}
              variant={item.buttonVariant}
              onClick={item.label === 'Save' ? onSubmit : item.function}
            >
              {item.label}
            </Button>
          ))}
        </ButtonGroup>
      </div>
    </>
  )
}

export default NodeEdit
