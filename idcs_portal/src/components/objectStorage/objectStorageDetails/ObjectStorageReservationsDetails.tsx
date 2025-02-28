// // INTEL CONFIDENTIAL
// // Copyright (C) 2023 Intel Corporation

import React from 'react'
import GridPagination from '../../../utils/gridPagination/gridPagination'
import TapContent from '../../../utils/TapContent/TapContent'
import { Dropdown, DropdownButton, ButtonGroup } from 'react-bootstrap'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import { NavLink, Link } from 'react-router-dom'
import { Plus } from 'react-bootstrap-icons'
import idcConfig from '../../../config/configurator'
import AfterAction from '../../../utils/modals/afterAction/AfterAction'
import TabsNavigation from '../../../utils/tabsNavigation/TabsNavagation'
import Spinner from '../../../utils/spinner/Spinner'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import { ReactComponent as ExternalLink } from '../../../assets/images/ExternalLink.svg'

interface ObjectStorageDetailsProps {
  reserveDetails: any
  bucketActiveTab: number
  loading: boolean
  tabs: any[]
  tabDetails: any
  reserveUser: any
  reserveLifeCyclePolicies: any
  userColumns: any[]
  policyColumns: any[]
  actionsReserveDetails: any[]
  actionModalContent: any
  showActionModal: boolean
  afterActionShowModal: boolean
  afterActionModalContent: any
  errorModalContent: any
  showErrorModal: boolean
  setBucketActiveTab: (value: number) => void
  setShowActionModal: (value: boolean) => Promise<void>
  setAction: (action: any, item: any) => void
  setAfterActionShowModal: (value: boolean) => void
  setShowErrorModal: (show: boolean) => void
}

const ObjectStorageReservationsDetails: React.FC<ObjectStorageDetailsProps> = (props): JSX.Element => {
  // props
  const reserveDetails = props.reserveDetails
  const bucketActiveTab = props.bucketActiveTab
  const loading = props.loading
  const tabs = props.tabs
  const tabDetails = props.tabDetails
  const securtyTabContent = props.reserveUser
  const policyTabContent = props.reserveLifeCyclePolicies
  const userColumns = props.userColumns
  const policyColumns = props.policyColumns
  const actionsReserveDetails = props.actionsReserveDetails
  const actionModalContent = props.actionModalContent
  const showActionModal = props.showActionModal
  const afterActionShowModal = props.afterActionShowModal
  const afterActionModalContent = props.afterActionModalContent
  const errorModalContent = props.errorModalContent
  const showErrorModal = props.showErrorModal
  const setBucketActiveTab: any = props.setBucketActiveTab
  const setShowActionModal: any = props.setShowActionModal
  const setAction = props.setAction
  const afterActionClickModal = props.setAfterActionShowModal
  const setShowErrorModal = props.setShowErrorModal

  // variables
  const spinner: JSX.Element = <Spinner />
  const instanceDetails: JSX.Element = (
    <>
      <div className="section flex-row bd-highlight">
        <h2>Bucket: {reserveDetails?.name}</h2>
        {reserveDetails?.status !== 'Failed' ? (
          <Link intc-id="btn how-to-connect" to={idcConfig.REACT_APP_GUIDES_OBJECT_STORAGE_URL} target="_blank">
            <span>How to access</span>
            <ExternalLink />
          </Link>
        ) : null}
        {actionsReserveDetails.length > 0 ? (
          <DropdownButton
            as={ButtonGroup}
            variant="outline-primary"
            title="Actions"
            intc-id="myReservationActionsDropdownButton"
          >
            {actionsReserveDetails.map((action: any, index: number) => (
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
        <TabsNavigation tabs={tabs} activeTab={bucketActiveTab} setTabActive={setBucketActiveTab} />
        <TapContent infoToDisplay={getTabInfo(bucketActiveTab)} />
      </div>
    </>
  )

  function getTabInfo(bucketActiveTab: number): any {
    let content = null
    switch (bucketActiveTab) {
      case 1:
        tabDetails[bucketActiveTab].customContent = getSecurityTabInfo()
        content = tabDetails[bucketActiveTab]
        break
      case 2:
        tabDetails[bucketActiveTab].customContent = getLifecycleRuleTabInfo()
        content = tabDetails[bucketActiveTab]
        break
      default:
        content = tabDetails[bucketActiveTab]
        break
    }
    return content
  }
  function getSecurityTabInfo(): JSX.Element {
    const content = (
      <>
        <hr />
        <div className="row">
          <div className="colxs-12 col-xl-10 col-xxl-8">
            <GridPagination data={securtyTabContent} columns={userColumns} />
          </div>
        </div>
      </>
    )
    return content
  }
  function getLifecycleRuleTabInfo(): JSX.Element {
    const content = (
      <>
        {policyTabContent && policyTabContent.length > 0 && (
          <GridPagination data={policyTabContent} columns={policyColumns} />
        )}
        <NavLink
          to={`/buckets/d/${reserveDetails?.name}/lifecyclerule/reserve`}
          className="btn btn-outline-primary"
          intc-id={'btn-create-bucket-rule'}
        >
          <Plus /> Create rule
        </NavLink>
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
      <AfterAction
        showModal={afterActionShowModal}
        modalContent={afterActionModalContent}
        onClickModal={afterActionClickModal}
      />
      {loading || !reserveDetails ? spinner : instanceDetails}
    </>
  )
}

export default ObjectStorageReservationsDetails
