// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { BsPlusLg } from 'react-icons/bs'
import GridPagination from '../../../utils/gridPagination/gridPagination'
import AddMemberView from './AddMemberView'
import RemoveMemberView from './RemoveMemberView'
import idcConfig from '../../../config/configurator'
import { Button } from 'react-bootstrap'
import SearchBox from '../../../utils/searchBox/SearchBox'
import { ReactComponent as ExternalLink } from '../../../assets/images/ExternalLink.svg'

const AccessManagementView = ({
  cloudAccountId,
  accessManagementColumns,
  accessManagementGridInfo,
  addNewMember,
  filterText,
  setFilter,
  invitationLoading,
  accountManagementEmptyGrid,
  accountManagementEmptyGridNoResults,
  addMemberForm,
  onChangeInput,
  removeMember,
  invitationLimit,
  onChangeDropdownMultiple,
  roles
}) => {
  const accountManagementGridData = accessManagementGridInfo.filter(
    (x) => !filterText || x.email.indexOf(filterText) !== -1
  )
  const gridFeedBack = `${accountManagementGridData.length} of ${invitationLimit} members invited`

  return (
    <>
      <AddMemberView
        cloudAccountId={cloudAccountId}
        addMemberForm={addMemberForm}
        onChangeInput={onChangeInput}
        onChangeDropdownMultiple={onChangeDropdownMultiple}
        roles={roles}
      />
      <RemoveMemberView removeMember={removeMember} />
      <div className="section">
        <h2>Account Access Management</h2>
        <span className="d-flex flex-row align-items-center">
          The following have permissions to manage your account resources.&nbsp;
          <a href={idcConfig.REACT_APP_MULTIUSER_GUIDE} target="_blank" rel="noreferrer" className="link">
            Learn about account access management
            <ExternalLink />
          </a>
        </span>
      </div>
      {accessManagementGridInfo && accessManagementGridInfo.length > 0 && (
        <div className="filter">
          <Button
            intc-id="btn-profile-addMember"
            data-wap_ref="btn-profile-addMember"
            variant="primary"
            disabled={accountManagementGridData.length >= invitationLimit}
            onClick={() => addNewMember()}
          >
            <BsPlusLg />
            Grant access
          </Button>
          <div className="d-flex justify-content-end">
            <SearchBox
              intc-id="btn-profile-saerchmember"
              value={filterText}
              onChange={setFilter}
              placeholder="Search members by email..."
              aria-label="Type an email to search a member.."
            />
          </div>
        </div>
      )}
      <div className="section">
        <GridPagination
          data={accountManagementGridData}
          columns={accessManagementColumns}
          emptyGrid={filterText ? accountManagementEmptyGridNoResults : accountManagementEmptyGrid}
          loading={invitationLoading}
          feedBack={gridFeedBack}
        />
      </div>
    </>
  )
}

export default AccessManagementView
