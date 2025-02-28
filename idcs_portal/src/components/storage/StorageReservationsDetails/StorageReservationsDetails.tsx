// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import TapContent from '../../../utils/TapContent/TapContent'
import { Dropdown, DropdownButton, ButtonGroup, Button } from 'react-bootstrap'
import HowToConnect from '../../../utils/modals/howToConnect/HowToConnect'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import { BsCopy } from 'react-icons/bs'
import TabsNavigation from '../../../utils/tabsNavigation/TabsNavagation'
import LabelValuePair from '../../../utils/labelValuePair/LabelValuePair'
import React from 'react'
import Spinner from '../../../utils/spinner/Spinner'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'

interface StorageReservationDetailsProps {
  reserveDetails: any
  activeTab: number
  tabDetails: any[]
  actionsReserveDetails: any[]
  showHowToMountModal: boolean
  showHowToUnmountModal: boolean
  showActionModal: boolean
  actionModalContent: any
  tabs: any[]
  loading: boolean
  showGeneratePwd: any
  userProfile: any
  errorModalContent: any
  showErrorModal: boolean
  setShowErrorModal: (show: boolean) => void
  setShowHowToMountModal: (value: boolean) => void
  setShowHowToUnmountModal: (value: boolean) => void
  setShowActionModal: (value: any) => void
  setActiveTab: (value: number) => void
  setAction: (action: any, item: any) => void
  generatePwd: () => Promise<void>
}

const StorageReservationsDetails: React.FC<StorageReservationDetailsProps> = (props): JSX.Element => {
  // props
  const reserveDetails = props.reserveDetails
  const activeTab = props.activeTab
  const tabDetails = props.tabDetails
  const actionsReserveDetails = props.actionsReserveDetails
  const showHowToMountModal = props.showHowToMountModal
  const showHowToUnmountModal = props.showHowToUnmountModal
  const showActionModal = props.showActionModal
  const actionModalContent = props.actionModalContent
  const tabs = props.tabs
  const loading = props.loading
  const showGeneratePwd = props.showGeneratePwd
  const userProfile = props.userProfile
  const errorModalContent = props.errorModalContent
  const showErrorModal = props.showErrorModal
  const setShowErrorModal = props.setShowErrorModal
  const setShowHowToMountModal = props.setShowHowToMountModal
  const setShowHowToUnmountModal = props.setShowHowToUnmountModal
  const setShowActionModal = props.setShowActionModal
  const setActiveTab: any = props.setActiveTab
  const setAction = props.setAction
  const generatePwd: any = props.generatePwd

  // variables
  const spinner = <Spinner />

  const instanceDetails = (
    <>
      <div className="section flex-row bd-highlight">
        <h2>My volume: {reserveDetails?.name}</h2>
        {reserveDetails?.status !== 'Failed' ? (
          <>
            <Button
              variant="outline-primary"
              intc-id="btn how-to-connect"
              onClick={() => {
                setShowHowToMountModal(true)
              }}
            >
              How to mount
            </Button>
            <Button
              variant="outline-primary"
              intc-id="btn how-to-connect"
              onClick={() => {
                setShowHowToUnmountModal(true)
              }}
            >
              How to unmount
            </Button>
          </>
        ) : null}
        {actionsReserveDetails && actionsReserveDetails?.length > 0 ? (
          <DropdownButton
            as={ButtonGroup}
            variant="outline-primary"
            title="Actions"
            intc-id="myReservationActionsDropdownButton"
          >
            {actionsReserveDetails.map((action, index) => (
              <Dropdown.Item
                key={index}
                onClick={() => {
                  setAction(action, reserveDetails)
                }}
                intc-id={`myReservationActionsDropdownItemButton${index}`}
              >
                {action.name}
              </Dropdown.Item>
            ))}
          </DropdownButton>
        ) : null}
      </div>
      <div className="section">
        <TabsNavigation tabs={tabs} activeTab={activeTab} setTabActive={setActiveTab} />
        <TapContent infoToDisplay={getTabInfo(activeTab)} />
      </div>
    </>
  )

  function getTabInfo(activeTab: number): any {
    let content = null
    switch (activeTab) {
      case 1:
        tabDetails[activeTab].customContent = getSecurityTabInfo(tabDetails[activeTab])
        content = tabDetails[activeTab]
        break
      default:
        content = tabDetails[activeTab]
        break
    }

    return content
  }

  function getSecurityTabInfo(tabContent: any): JSX.Element {
    const { fields } = tabContent
    const content = (
      <>
        {fields.map((field: any, index: number) => (
          <div className="d-flex flex-column gap-s6" key={index}>
            {field.field === 'user' || userProfile ? (
              <LabelValuePair label={field.label}>
                <>
                  {field.mask ? field.value.replace(/./g, '*') : field.value}
                  {field.action
                    ? field.action.map((item: any, index: number) => (
                        <Button
                          variant="icon-simple"
                          aria-label={`Copy ${field.label.replace(':', '')}`}
                          onClick={() => item.func(field.value)}
                          key={index}
                        >
                          <BsCopy />
                        </Button>
                      ))
                    : null}
                </>
              </LabelValuePair>
            ) : (
              <Button variant="outline-primary" size="sm" onClick={generatePwd}>
                {showGeneratePwd.icon}&nbsp;
                {showGeneratePwd.label}
              </Button>
            )}
          </div>
        ))}
      </>
    )

    return content
  }

  return (
    <>
      <ErrorModal
        showModal={showErrorModal}
        titleMessage={errorModalContent.titleMessage}
        description={errorModalContent.description}
        message={errorModalContent.message}
        onClickCloseErrorModal={() => {
          setShowErrorModal(false)
        }}
      />
      <ActionConfirmation
        actionModalContent={actionModalContent}
        onClickModalConfirmation={setShowActionModal}
        showModalActionConfirmation={showActionModal}
      />
      <HowToConnect
        flowType="mount"
        data={reserveDetails}
        showHowToConnectModal={showHowToMountModal}
        onClickHowToConnect={setShowHowToMountModal}
      />
      <HowToConnect
        flowType="unmount"
        data={reserveDetails}
        showHowToConnectModal={showHowToUnmountModal}
        onClickHowToConnect={setShowHowToUnmountModal}
      />
      {loading || !reserveDetails ? spinner : instanceDetails}
    </>
  )
}

export default StorageReservationsDetails
