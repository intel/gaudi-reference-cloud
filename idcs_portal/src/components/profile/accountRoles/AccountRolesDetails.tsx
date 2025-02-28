// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { useNavigate } from 'react-router-dom'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import Spinner from '../../../utils/spinner/Spinner'
import { DropdownButton, ButtonGroup, Dropdown, Accordion, Card } from 'react-bootstrap'
import TapContent from '../../../utils/TapContent/TapContent'
import TabsNavigation from '../../../utils/tabsNavigation/TabsNavagation'
import GridPagination from '../../../utils/gridPagination/gridPagination'
import CustomInput from '../../../utils/customInput/CustomInput'
import PermissionHandler from './PermissionHandler'

interface UserRolesDetailsProps {
  isPageReady: boolean
  tabs: any[]
  reserveDetails: any
  tabDetails: any
  loading: boolean
  activeTab: number | string
  actionsReserveDetails: any[]
  showActionModal: boolean
  actionModalContent: any
  resources: any
  resourcesList: any
  adminInvitation: any
  emptyGrid: any
  form: any
  servicePermissions: any
  resourcePermissions: any
  isOwnCloudAccount: boolean
  setActiveTab: (tab: number | string) => void
  setAction: (action: any, item: any) => void
  setShowActionModal: (result: boolean) => Promise<void>
}

const AccountRolesDetails: React.FC<UserRolesDetailsProps> = (props): JSX.Element => {
  const navigate = useNavigate()

  const tabs = props.tabs
  const reserveDetails = props.reserveDetails
  const tabDetails = props.tabDetails
  const activeTab = props.activeTab
  const actionsReserveDetails = props.actionsReserveDetails
  const showActionModal = props.showActionModal
  const actionModalContent = props.actionModalContent
  const emptyGrid = props.emptyGrid
  const form = props.form
  const setActiveTab = props.setActiveTab
  const setAction = props.setAction
  const setShowActionModal: any = props.setShowActionModal
  const resources: any = props.resources
  const resourcesList: any = props.resourcesList
  const isPageReady: any = props.isPageReady
  const adminInvitation: any = props.adminInvitation
  const servicePermissions = props.servicePermissions
  const resourcePermissions = props.resourcePermissions
  const isOwnCloudAccount = props.isOwnCloudAccount

  const getPermissionTapInfo = (): JSX.Element => {
    if (reserveDetails && isPageReady) {
      return getResources()
    } else {
      return <></>
    }
  }

  function getTabInfo(activeTap: number | string): JSX.Element {
    let content = null
    switch (activeTap) {
      case 0:
        tabDetails[activeTap].customContent = getPermissionTapInfo()
        content = tabDetails[activeTap]
        break
      case 1:
        tabDetails[activeTap].customContent = getUsersTapInfo(tabDetails[activeTap])
        content = tabDetails[activeTap]
        break
      default:
        content = tabDetails[activeTap]
        break
    }

    return content
  }

  function getUsersTapInfo(tabContent: any): JSX.Element {
    const fields = tabContent.fields
    const columns = [
      {
        columnName: 'Assigned Users',
        targetColumn: 'users',
        columnConfig: {
          behaviorType: 'hyperLink',
          behaviorFunction: 'setDetails'
        }
      }
    ]

    const allowedStatus = ['INVITE_STATE_ACCEPTED']
    const acceptedInvites = adminInvitation
      .filter((x: any) => allowedStatus.includes(x.invitation_state))
      .map((x: any) => x.member_email)

    const data = fields[0].value.map((x: any) => {
      return {
        users: !acceptedInvites.includes(x)
          ? x
          : {
              showField: true,
              type: 'HyperLink',
              value: x,
              function: () => {
                navigate(`/profile/accountAccessManagement/user-role/${x}`)
              }
            }
      }
    })
    return (
      <div className="row">
        <div className={data.length > 0 ? 'col-xs-12 col-md-5 col-xl-4' : ''}>
          <GridPagination data={data} columns={columns} emptyGrid={emptyGrid} hidePaginationControl hideSortControls />
        </div>
      </div>
    )
  }

  const getResources = (): JSX.Element => {
    const items = resources.map((resource: any) => {
      if (resource.actions.length === 0) {
        return null
      }

      const perTypes = reserveDetails.permissions
        .map((x: any) => x.resourceType)
        .filter((value: string, index: number, array: any) => array.indexOf(value) === index)

      if (!perTypes.includes(resource.type)) {
        return (
          <Accordion key={resource.type} className="border-bottom-0">
            <Card key={resource.type} className="shadow-none">
              <Card.Header intc-id={`userRole-DetailsServiceTitle-${resource.type}`}>
                <h3>{resource.description}</h3>
              </Card.Header>
            </Card>
          </Accordion>
        )
      }

      return (
        <Accordion key={resource.type}>
          <Accordion.Item eventKey={resource.type} className="w-100 border-top">
            <Accordion.Header intc-id={`userRole-DetailsServiceTitle-${resource.type}`}>
              <h3>{resource.description}</h3>
            </Accordion.Header>
            <Accordion.Body className="d-flex flex-column gap-s4">
              <PermissionHandler
                resource={resource}
                form={form}
                servicePermissions={servicePermissions}
                resourcePermissions={resourcePermissions}
                resourcesList={resourcesList}
                isOwnCloudAccount={isOwnCloudAccount}
                viewMode={true}
                buildCustomInput={buildCustomInput}
              />
            </Accordion.Body>
          </Accordion.Item>
        </Accordion>
      )
    })

    return (
      <div className="section">
        <Accordion intc-id={'userRole-Details-accordion-'} className="d-flex-customInput border-0" alwaysOpen>
          {items}
        </Accordion>
      </div>
    )
  }

  const setReadyOnlyMode = (configInput: any): boolean => {
    if (configInput.type === 'checkbox') return !configInput.isChecked
    return configInput.isReadOnly
  }

  const buildCustomInput = (element: any): JSX.Element => {
    return (
      <CustomInput
        type={element.configInput.type}
        fieldSize={element.configInput.fieldSize}
        placeholder={element.configInput.placeholder}
        isRequired={element.configInput.validationRules.isRequired}
        label={
          element.configInput.validationRules.isRequired
            ? String(element.configInput.label) + ' *'
            : String(element.configInput.label)
        }
        value={element.configInput.value}
        onChanged={(event) => {}}
        isValid={element.configInput.isValid}
        isTouched={element.configInput.isTouched}
        helperMessage={element.configInput.helperMessage}
        isReadOnly={setReadyOnlyMode(element.configInput)}
        validationMessage={element.configInput.validationMessage}
        maxLength={element.configInput.maxLength}
        options={element.configInput.options}
        hidden={element.configInput.hidden}
        maxWidth={element.configInput.maxWidth}
        hiddenLabel={element.configInput.hiddenLabel}
        radioGroupHorizontal={element.configInput.radioGroupHorizontal}
        onChangeDropdownMultiple={(values) => {}}
      />
    )
  }

  const spinner = <Spinner />

  const roleDetails = (
    <>
      <div className="section flex-row bd-highlight">
        <h2>Role: {reserveDetails?.alias}</h2>

        {isOwnCloudAccount && actionsReserveDetails.length > 0 ? (
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
      {!isOwnCloudAccount && getPermissionTapInfo()}
      {isOwnCloudAccount && (
        <div className="section">
          <TabsNavigation tabs={tabs} activeTab={activeTab} setTabActive={setActiveTab} />
          <TapContent infoToDisplay={getTabInfo(activeTab)} />
        </div>
      )}
    </>
  )
  return (
    <>
      <ActionConfirmation
        actionModalContent={actionModalContent}
        onClickModalConfirmation={setShowActionModal}
        showModalActionConfirmation={showActionModal}
      />
      {!isPageReady ? spinner : roleDetails}
    </>
  )
}

export default AccountRolesDetails
