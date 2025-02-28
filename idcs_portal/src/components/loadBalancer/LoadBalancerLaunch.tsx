// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import CustomInput from '../../utils/customInput/CustomInput'
import ReservationSubmit from '../../utils/modals/reservationSubmit/ReservationSubmit'
import ErrorModal from '../../utils/modals/errorModal/ErrorModal'
import UpgradeNeededModal from '../../utils/modals/upgradeNeededModal/UpgradeNeededModal'
import { Button, ButtonGroup } from 'react-bootstrap'
import { BsNodePlus } from 'react-icons/bs'
import EmptyCatalogModal from '../../utils/modals/emptyCatalogModal/EmptyCatalogModal'

const LoadBalancerLaunch = (props: any): JSX.Element => {
  // props
  const state = props.state
  const form = props.form
  const onChangeInput = props.onChangeInput
  const onSubmit = props.onSubmit
  const onClickCloseErrorModal = props.onClickCloseErrorModal
  const showUpgradeNeededModal = props.showUpgradeNeededModal
  const setShowUpgradeNeededModal = props.setShowUpgradeNeededModal
  const onClickFormAction = props.onClickFormAction
  const onChangeTagValue = props.onChangeTagValue
  const onClickActionTag = props.onClickActionTag
  const onChangeDropdownMultiple = props.onChangeDropdownMultiple
  const selectAllInstances = props.selectAllInstances
  const emptyCatalogModal = props.emptyCatalogModal
  const onClickSourceIpsAction = props.onClickSourceIpsAction
  const maxListeners = props.maxListeners
  const maxSourceIps = props.maxSourceIps

  // State Variables
  const navigationBottom = state.navigationBottom
  const showReservationModal = state.showReservationModal
  const showErrorModal = state.showErrorModal
  const errorMessage = state.errorMessage

  // functions

  const buildCustomInput = (
    element: any,
    index: number | null = null,
    listenersItem: string | null = null
  ): JSX.Element => (
    <CustomInput
      key={element.id ?? '0'}
      type={element.configInput.type}
      placeholder={element.configInput.placeholder}
      isRequired={element.configInput.validationRules.isRequired}
      label={
        element.configInput.validationRules.isRequired
          ? String(element.configInput.label) + ' *'
          : element.configInput.label
      }
      value={element.configInput.value}
      onChanged={(event: any) => onChangeInput(event, element.id, index, listenersItem)}
      isValid={element.configInput.isValid}
      isTouched={element.configInput.isTouched}
      helperMessage={element.configInput.helperMessage}
      isReadOnly={element.configInput.isReadOnly}
      validationMessage={element.configInput.validationMessage}
      maxLength={element.configInput.maxLength}
      prepend={element.configInput.prependIcon}
      options={element.configInput.options}
      hidden={element.configInput.hidden}
      emptyOptionsMessage={element.configInput.emptyOptionsMessage}
      borderlessDropdownMultiple={element.configInput.borderlessDropdownMultiple}
      onChangeDropdownMultiple={(values: any) => onChangeDropdownMultiple(values, index, listenersItem)}
      selectAllButton={{
        label: 'Select All',
        buttonFunction: () => selectAllInstances(index)
      }}
      onChangeTagValue={(event: any, formInputName: string, tagIndex: number) =>
        onChangeTagValue(event, formInputName, tagIndex, index)
      }
      onClickActionTag={(tagIndex: any, actionType: string) => onClickActionTag(actionType, index, tagIndex)}
      dictionaryOptions={element.configInput.dictionaryOptions}
      maxWidth={element.configInput.maxWidth}
      radioGroupHorizontal={element.configInput.radioGroupHorizontal}
    />
  )

  const buildForm = (): any[] => {
    const formInputs = []

    const listenersItems = form.listeners.items

    for (const key in listenersItems) {
      const item = { ...listenersItems[key] }
      const mainDiv = (
        <div className="d-flex flex-column w-100 gap-s6" key={key}>
          <div className="d-flex flex-row gap-s8 align-items-center">
            <h4>Listener</h4>
            <Button
              intc-id="btn-loadBalancer-delete-listener"
              data-wap_ref="btn-loadBalancer-delete-listener"
              variant="close"
              aria-label="Delete Listener"
              onClick={() => onClickFormAction('Delete', key)}
            ></Button>
          </div>
          <div className="d-flex flex-xs-column flex-md-row gap-s6">
            <div className="d-flex flex-row gap-s6 w-auto">{buildListeners(item, 'listeners1', key)}</div>
            <div className="d-flex flex-xs-column flex-sm-row w-auto gap-s6 flex-grow-1">
              {buildListeners(item, 'listeners2', key)}
            </div>
          </div>
          {buildListeners(item, 'pool', key)}
          <hr />
        </div>
      )
      formInputs.push(mainDiv)
    }
    return formInputs
  }

  const buildIps = (): any[] => {
    const formInputs: any[] = []

    const ips = form.ips.items

    for (const key in ips) {
      const item = { ...ips[key] }

      const element = {
        id: 'ip',
        configInput: item.ip
      }

      const mainDiv = (
        <React.Fragment key={key}>
          <div className="d-flex w-100 gap-s6 align-items-start">
            {buildCustomInput(element, Number(key))}
            <Button
              intc-id="btn-loadBalancer-delete-source-ip"
              data-wap_ref="btn-loadBalancer-delete-source-ip"
              variant="close"
              aria-label="Delete Source IP"
              onClick={() => onClickSourceIpsAction('Delete', key)}
            ></Button>
          </div>
          <hr />
        </React.Fragment>
      )
      formInputs.push(mainDiv)
    }

    return formInputs
  }

  const buildListeners = (listenersItem: any, sectionGroup: string, index: number | string): any[] => {
    const formInputs = []

    for (const key in listenersItem) {
      const item = { ...listenersItem[key] }
      if (item.sectionGroup === sectionGroup) {
        const element = {
          id: key,
          configInput: item
        }
        formInputs.push(
          <React.Fragment key={key}>{buildCustomInput(element, Number(index), 'listeners')}</React.Fragment>
        )
      }
    }
    return formInputs
  }

  const content = (
    <>
      <div className="section">
        <h2 intc-id="LoadBalancerLaunchTitle">Launch a Load Balancer</h2>
      </div>
      <div className="section">
        {buildCustomInput({ id: 'name', configInput: form.name })}

        <h3 intc-id="LoadBalancerSourceIpsTitle">Source IPs</h3>
        {buildIps()}
        <div className="d-flex flex-row gap-s6 align-items-center">
          <Button
            intc-id="btn-loadBalancer-addSourceIP"
            data-wap_ref="btn-loadBalancer-addSourceIP"
            aria-label="Add Source IP"
            variant="outline-primary"
            onClick={() => onClickSourceIpsAction('Add')}
            disabled={maxSourceIps !== '' && form.ips.items.length >= maxSourceIps}
          >
            <BsNodePlus />
            Add Source IP
          </Button>
          <div className="d-flex flex-row w-100 gap-s8 justify-content-between">
            {maxSourceIps !== '' && (
              <>{`Up to ${maxSourceIps} source IPs max. (${maxSourceIps - form.ips.items.length}) remaining`}</>
            )}
          </div>
        </div>

        <h3 intc-id="LoadBalancerListenersTitle">Listeners</h3>
        {buildForm()}
        <div className="d-flex flex-row gap-s6 align-items-center">
          <Button
            intc-id="btn-loadBalancer-addListener"
            data-wap_ref="btn-loadBalancer-addListener"
            aria-label="Add Listener"
            variant="outline-primary"
            onClick={() => onClickFormAction('Add')}
            disabled={maxListeners !== '' && form.listeners.items.length >= maxListeners}
          >
            <BsNodePlus />
            Add Listener
          </Button>

          <div className="d-flex flex-row w-100 gap-s8 justify-content-between">
            {maxListeners !== '' && (
              <>{`Up to ${maxListeners} listeners max. (${maxListeners - form.listeners.items.length}) remaining`}</>
            )}
          </div>
        </div>

        <ButtonGroup>
          {navigationBottom.map(
            (item: any, index: number): JSX.Element => (
              <Button
                intc-id={'btn-LoadBalancer-navigationBottom' + String(item.buttonLabel)}
                data-wap_ref={'btn-LoadBalancer-navigationBottom' + String(item.buttonLabel)}
                aria-label={item.buttonLabel}
                key={index}
                variant={item.buttonVariant}
                onClick={item.buttonLabel === 'Launch' ? onSubmit : item.buttonFunction}
              >
                {item.buttonLabel}
              </Button>
            )
          )}
        </ButtonGroup>
      </div>
    </>
  )

  return (
    <>
      <ReservationSubmit showReservationCreateModal={showReservationModal} />
      <ErrorModal
        showModal={showErrorModal}
        titleMessage="Could not create your load balancer"
        description={'There was an error while processing your load balancer.'}
        message={errorMessage}
        onClickCloseErrorModal={onClickCloseErrorModal}
      />
      <UpgradeNeededModal showModal={showUpgradeNeededModal} onClose={() => setShowUpgradeNeededModal(false)} />
      <EmptyCatalogModal show={emptyCatalogModal} product="Load Balancer" goBackPath="/load-balancer" />
      {content}
    </>
  )
}

export default LoadBalancerLaunch
