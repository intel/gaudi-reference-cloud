// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporationimport React from 'react'

import React from 'react'
import GridPagination from '../../../utils/gridPagination/gridPagination'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import CustomInput from '../../../utils/customInput/CustomInput'
import { Button, Spinner } from 'react-bootstrap'
import SearchBox from '../../../utils/searchBox/SearchBox'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'

interface AccessManagementEditRolesProps {
  columns: any[]
  loading: boolean
  isPageReady: boolean
  isAssignRole: boolean
  isRemoveRole: boolean
  emptyGrid: any
  userRoles: any[]
  user: string | undefined
  form: any
  roles: any[]
  showActionModal: boolean
  actionModalContent: any
  filterText: string
  errorModal: any
  setFilter: (event: any, clear: boolean) => void
  actionOnModal: (result: boolean) => Promise<void>
  onChangeInput: (event: any, formInputName: string) => void
  onChangeDropdownMultiple: (values: string[] | [], formInputName: string) => void
  addUserToRole: (event: any) => Promise<void>
  onClickCloseErrorModal: () => void
}

const UserRoles: React.FC<AccessManagementEditRolesProps> = (props): JSX.Element => {
  const columns = props.columns
  const loading = props.loading
  const isPageReady = props.isPageReady
  const isAssignRole = props.isAssignRole
  const isRemoveRole = props.isRemoveRole
  const emptyGrid = props.emptyGrid
  const userRoles = props.userRoles
  const user = props.user
  const form = props.form
  const showActionModal = props.showActionModal
  const actionModalContent = props.actionModalContent
  const filterText = props.filterText
  const setFilter: any = props.setFilter
  const actionOnModal: any = props.actionOnModal
  const onChangeInput = props.onChangeInput
  const addUserToRole: any = props.addUserToRole
  const onChangeDropdownMultiple = props.onChangeDropdownMultiple
  const errorModal = props.errorModal
  const onClickCloseErrorModal = props.onClickCloseErrorModal

  // variables
  let gridItems = []

  if (filterText !== '') {
    const input = filterText.toLowerCase()

    gridItems = userRoles.filter((item: any) =>
      item.roles.value ? item.roles.value.toLowerCase().includes(input) : item.roles.toLowerCase().includes(input)
    )
  } else {
    gridItems = userRoles
  }

  // functions
  function buildCustomInput(element: any): JSX.Element {
    return (
      <CustomInput
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
        onChanged={(event) => {
          onChangeInput(event, element.id)
        }}
        isValid={element.configInput.isValid}
        isTouched={element.configInput.isTouched}
        helperMessage={element.configInput.helperMessage}
        isReadOnly={element.configInput.isReadOnly}
        validationMessage={element.configInput.validationMessage}
        maxLength={element.configInput.maxLength}
        maxWidth={element.configInput.maxWidth}
        options={element.configInput.options}
        prepend={element.configInput.prepend}
        hiddenLabel={element.configInput.hiddenLabel}
        onChangeDropdownMultiple={(values) => {
          onChangeDropdownMultiple(values, element.id)
        }}
        emptyOptionsMessage={element.configInput.emptyOptionsMessage}
      />
    )
  }

  return !isPageReady ? (
    <Spinner />
  ) : (
    <>
      <div className="section mb-sa9">
        <h2>Assign role to {user}</h2>

        <div className="d-flex w-100 gap-s6">
          {buildCustomInput({
            id: 'role',
            configInput: form.role
          })}
          <div>
            <Button
              variant="primary"
              intc-id="btn-add-user-to-role"
              data-wap_ref="btn-add-user-to-role"
              onClick={addUserToRole}
              disabled={form.role?.options?.length === 0 || isAssignRole}
            >
              {isAssignRole ? <Spinner as="span" size="sm" /> : 'Add role to user'}
            </Button>
          </div>
        </div>
      </div>
      <div className="filter">
        <h3>User Roles</h3>
        {gridItems.length > 0 && (
          <SearchBox
            intc-id="searchRoles"
            value={filterText}
            onChange={setFilter}
            placeholder="Search roles..."
            aria-label="Type to search role.."
          />
        )}
      </div>
      <GridPagination data={gridItems} columns={columns} loading={loading || isRemoveRole} emptyGrid={emptyGrid} />
      <ActionConfirmation
        actionModalContent={actionModalContent}
        onClickModalConfirmation={actionOnModal}
        showModalActionConfirmation={showActionModal}
      />
      <ErrorModal
        showModal={errorModal.showErrorModal}
        titleMessage={errorModal.errorTitleMessage}
        description={errorModal.errorDescription}
        message={errorModal.errorMessage}
        hideRetryMessage={errorModal.errorHideRetryMessage}
        onClickCloseErrorModal={onClickCloseErrorModal}
      />
    </>
  )
}
export default UserRoles
