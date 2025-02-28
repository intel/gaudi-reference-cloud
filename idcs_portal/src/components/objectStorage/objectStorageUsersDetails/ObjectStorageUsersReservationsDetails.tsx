// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import TapContent from '../../../utils/TapContent/TapContent'
import { Dropdown, DropdownButton, ButtonGroup, Button } from 'react-bootstrap'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import { BsCopy } from 'react-icons/bs'
import ObjectStorageUsersPermissionsManagement from '../objectStorageUsersPermissionsManagement/ObjectStorageUsersPermissionsManagement'
import TabsNavigation from '../../../utils/tabsNavigation/TabsNavagation'
import LabelValuePair from '../../../utils/labelValuePair/LabelValuePair'
import CustomAlerts from '../../../utils/customAlerts/CustomAlerts'
import idcConfig from '../../../config/configurator'
import Spinner from '../../../utils/spinner/Spinner'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'

interface ObjectStorageUsersDetailsProps {
  reserveDetails: any
  activeTab: number
  tabDetails: any[]
  actionsReserveDetails: any[]
  showActionModal: boolean
  actionModalContent: any
  tabs: any[]
  loading: boolean
  showGeneratePwd: any
  showAccessKey: string | null
  showSecretKey: string | null
  userBuckets: any[]
  errorModalContent: any
  showErrorModal: boolean
  setShowErrorModal: (value: boolean) => void
  setShowActionModal: (value: boolean) => void
  setActiveTab: (value: number) => void
  setAction: (action: any, item: any) => void
  generatePwd: () => Promise<void>
  copyItem: (value: any) => void
}

const ObjectStorageUsersReservationsDetails: React.FC<ObjectStorageUsersDetailsProps> = (props) => {
  // props
  const reserveDetails = props.reserveDetails
  const activeTab = props.activeTab
  const tabDetails = props.tabDetails
  const actionsReserveDetails = props.actionsReserveDetails
  const showActionModal = props.showActionModal
  const actionModalContent = props.actionModalContent
  const tabs = props.tabs
  const loading = props.loading
  const showGeneratePwd = props.showGeneratePwd
  const showAccessKey = props.showAccessKey
  const showSecretKey = props.showSecretKey
  const userBuckets = props.userBuckets
  const errorModalContent = props.errorModalContent
  const showErrorModal = props.showErrorModal
  const setShowErrorModal = props.setShowErrorModal
  const setShowActionModal = props.setShowActionModal
  const setActiveTab: any = props.setActiveTab
  const setAction = props.setAction
  const generatePwd: any = props.generatePwd
  const copyItem = props.copyItem

  // variables
  const spinner = <Spinner />

  const instanceDetails = (
    <>
      <div className="section flex-row bd-highlight">
        <h2>Principal: {reserveDetails?.name}</h2>
        {actionsReserveDetails.length > 0 ? (
          <DropdownButton
            as={ButtonGroup}
            variant="outline-primary"
            title="Actions"
            intc-id="myUsersActionsDropdownButton"
          >
            {actionsReserveDetails.map((action, index) => (
              <Dropdown.Item
                key={index}
                onClick={() => {
                  setAction(action, reserveDetails)
                }}
                intc-id={`myUsersActionsDropdownItemButton${index}`}
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
        tabDetails[activeTab].customContent = getPermissionTabInfo()
        content = tabDetails[activeTab]
        break
      default:
        tabDetails[activeTab].customContent = getCredentialsTabInfo(tabDetails[activeTab])
        content = tabDetails[activeTab]
        break
    }

    return content
  }

  function getCredentialsTabInfo(tabContent: any): JSX.Element {
    const { fields } = tabContent
    const content = (
      <>
        <CustomAlerts
          showAlert={true}
          alertType="secondary"
          message={`The credentials are generated once, ${idcConfig.REACT_APP_CONSOLE_SHORT_NAME} doesn’t store passwords.`}
          onCloseAlert={() => {}}
          showIcon={true}
        />
        <div className="row">
          {fields.map((field: any, index: number) => (
            <LabelValuePair className="col-md-3" label={field.label} key={index}>
              {field.field === 'name' ? (
                <>
                  {field.value}
                  {field.action
                    ? field.action.map((item: any, index: number) => (
                        <Button
                          onClick={() => item.func(field.value)}
                          key={index}
                          variant="icon-simple"
                          aria-label="Copy principal"
                        >
                          <BsCopy />
                        </Button>
                      ))
                    : null}
                  <Button aria-label={showGeneratePwd.label} size="sm" variant="outline-primary" onClick={generatePwd}>
                    {showGeneratePwd.icon}&nbsp;
                    {showGeneratePwd.label}
                  </Button>
                </>
              ) : (
                <></>
              )}
            </LabelValuePair>
          ))}
        </div>
        {showAccessKey && showSecretKey && (
          <div className="d-flex flex-column gap-s6">
            <LabelValuePair label="Access Key:">
              <>
                {showAccessKey}
                <Button
                  aria-label="Copy access key:"
                  variant="icon-simple"
                  onClick={() => {
                    copyItem(showAccessKey)
                  }}
                >
                  <BsCopy />
                </Button>
              </>
            </LabelValuePair>
            <LabelValuePair label="Secret Key:">
              <>
                {showSecretKey}
                <Button
                  aria-label="Copy secret key"
                  variant="icon-simple"
                  onClick={() => {
                    copyItem(showSecretKey)
                  }}
                >
                  <BsCopy />
                </Button>
              </>
            </LabelValuePair>
          </div>
        )}
      </>
    )

    return content
  }

  function getPermissionTabInfo(): JSX.Element {
    const content = (
      <div className="section px-0">
        <div className="d-flex flex-xs-column flex-md-row gap-s8">
          {userBuckets && userBuckets.length > 0 && (
            <ObjectStorageUsersPermissionsManagement isView={true} buckets={userBuckets} />
          )}
        </div>
      </div>
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
      {loading || !reserveDetails ? spinner : instanceDetails}
    </>
  )
}

export default ObjectStorageUsersReservationsDetails
