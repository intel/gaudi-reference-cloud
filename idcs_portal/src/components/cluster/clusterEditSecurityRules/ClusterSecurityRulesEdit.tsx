// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import CustomInput from '../../../utils/customInput/CustomInput'
import Button from 'react-bootstrap/Button'
import { ButtonGroup } from 'react-bootstrap'
import { BsNodePlusFill } from 'react-icons/bs'
import ReservationSubmit from '../../../utils/modals/reservationSubmit/ReservationSubmit'

const ClusterSecurityRulesEdit = (props: any): JSX.Element => {
  // *****
  // Variables
  // *****
  const state = props.state
  const mainTitle = state.mainTitle
  const form = state.form
  const onChangeInput = props.onChangeInput
  const navigationBottom = state.navigationBottom
  const onSubmit = props.onSubmit
  const onClickActionSourceIp = props.onClickActionSourceIp
  const sourceIpsLimit = props.sourceIpsLimit
  const submitModal = props.submitModal

  const formElementsSourceIpSection = []
  const formElementsPortSection = []

  for (const key in form) {
    const formItem = {
      ...form[key]
    }

    if (key === 'ips') {
      const sourceIps = formItem.items
      for (const index in sourceIps) {
        const item = { ...sourceIps[index] }
        const sourceIpElements: any = []
        for (const itemkey in item) {
          const itemElement = { ...item[itemkey] }
          sourceIpElements.push({
            id: itemkey,
            idParent: key,
            nodeIndex: index,
            configInput: itemElement
          })
        }
        formElementsSourceIpSection.push({
          id: key,
          items: sourceIpElements
        })
      }
    }

    if (formItem.sectionGroup === 'port') {
      formElementsPortSection.push({
        id: key,
        configInput: formItem
      })
    }
  }

  // *****
  // Functions
  // *****
  const buildCustomInput = (element: any, key: number): JSX.Element => {
    return (
      <CustomInput
        key={key}
        type={element.configInput.type}
        fieldSize={element.configInput.fieldSize}
        placeholder={element.configInput.placeholder}
        isRequired={element.configInput.validationRules.isRequired}
        label={
          element.configInput.validationRules.isRequired
            ? String(element.configInput.label) + ' *'
            : element.configInput.label
        }
        value={element.configInput.value}
        onChanged={(event) => onChangeInput(event, element.id, element.idParent, element.nodeIndex)}
        isValid={element.configInput.isValid}
        isTouched={element.configInput.isTouched}
        helperMessage={element.configInput.helperMessage}
        isReadOnly={element.configInput.isReadOnly}
        options={element.configInput.options}
        validationMessage={element.configInput.validationMessage}
        refreshButton={element.configInput.refreshButton}
        extraButton={element.configInput.extraButton}
        selectAllButton={element.configInput.selectAllButton}
        labelButton={element.configInput.labelButton}
        emptyOptionsMessage={element.configInput.emptyOptionsMessage}
        hidden={element.configInput.hidden}
      />
    )
  }

  return (
    <>
      <ReservationSubmit showReservationCreateModal={submitModal.show} />
      <div className="section">
        <h2 intc-id="computeReserveTitle">{mainTitle}</h2>
      </div>
      <div className="section">
        {formElementsSourceIpSection.map((element, index) => {
          return (
            <React.Fragment key={index}>
              <div className="d-flex flex-row gap-s6 w-100 align-items-start">
                {element.items.map((item: any) => {
                  return buildCustomInput(item, index)
                })}
                <Button
                  onClick={() => onClickActionSourceIp(index, 'Delete')}
                  disabled={index === 0}
                  variant="close"
                  type="button"
                  aria-label="Close"
                  intc-id="btn-IKS-delete-source-ip"
                  data-wap_ref="btn-IKS-delete-source-ip"
                ></Button>
              </div>
              <hr />
            </React.Fragment>
          )
        })}
        <ButtonGroup>
          <Button
            intc-id="btn-iksMyClusters-addSourceIp"
            data-wap_ref="btn-iksMyClusters-addSourceIp"
            variant="outline-primary"
            disabled={formElementsSourceIpSection.length >= sourceIpsLimit}
            onClick={() => onClickActionSourceIp(null, 'Add')}
          >
            <BsNodePlusFill />
            Add Source Ip
          </Button>
          <span className="feedback">
            {formElementsSourceIpSection.length} of {sourceIpsLimit} Source Ips
          </span>
        </ButtonGroup>
        {formElementsPortSection.map((element, index) => buildCustomInput(element, index))}
        <ButtonGroup>
          {navigationBottom.map((item: any, index: string) => (
            <Button
              intc-id={`btn-ikscluster-ruleEdit-navigationBottom ${item.buttonLabel}`}
              data-wap_ref={`btn-ikscluster-ruleEdit-navigationBottom ${item.buttonLabel}`}
              aria-label={item.buttonLabel}
              key={index}
              variant={item.buttonVariant}
              onClick={item.buttonAction === 'Submit' ? onSubmit : item.buttonFunction}
            >
              {item.buttonLabel}
            </Button>
          ))}
        </ButtonGroup>
      </div>
    </>
  )
}

export default ClusterSecurityRulesEdit
