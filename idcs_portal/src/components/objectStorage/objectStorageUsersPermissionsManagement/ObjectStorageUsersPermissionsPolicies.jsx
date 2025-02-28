// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import useBucketUsersPermissionsStore from '../../../store/bucketUsersPermissionsStore/BucketUsersPermissionsStore'
import { objectStorageUsersPermissionPolicies } from '../../../utils/Enums'
import CustomInput from '../../../utils/customInput/CustomInput'

const ObjectStorageUsersPermissionsPolicies = (props) => {
  // props
  const isView = props.isView || false
  const bucketId = props.bucketId

  const bucketsPermissions = useBucketUsersPermissionsStore((state) => state.bucketsPermissions)
  const setPermissions = useBucketUsersPermissionsStore((state) => state.setPermissions)
  const setPrefix = useBucketUsersPermissionsStore((state) => state.setPrefix)

  const selectedBucket = bucketsPermissions !== null ? bucketsPermissions[bucketId] : null
  const permissions = selectedBucket?.permission ?? []
  const prefix = selectedBucket?.prefix ?? ''

  const onSelectAllPermissions = () => {
    const objectKeys = Object.keys(objectStorageUsersPermissionPolicies)
    const selectedValues = permissions
    const shouldDeselect = objectKeys.every((x) => selectedValues.includes(x))
    setPermissions(bucketId, shouldDeselect ? [] : objectKeys)
  }

  const clickMultipleSelectValues = (selectedOptions) => {
    setPermissions(bucketId, selectedOptions)
  }

  const getAllowedPolicies = () => {
    if (isView) {
      return (
        <div className="d-flex flex-column gap-s4 text-nowrap">
          <label className="mb-0 form-label">Allowed policies:</label>
          {permissions.map((item, index) => (
            <div key={index}>
              <span>{item}</span>
            </div>
          ))}
        </div>
      )
    } else {
      const allowedPoliciesConfig = {
        label: 'Allowed policies:',
        type: 'multi-select',
        customClass: 'text-nowrap',
        borderlessDropdownMultiple: true,
        onChangeDropdownMultiple: clickMultipleSelectValues,
        options: [],
        selectAllButton: {
          label: 'Select/Deselect All',
          buttonFunction: () => onSelectAllPermissions()
        }
      }

      for (const index in objectStorageUsersPermissionPolicies) {
        const name = objectStorageUsersPermissionPolicies[index]
        allowedPoliciesConfig.options.push({
          name,
          value: index
        })
      }
      return <CustomInput {...allowedPoliciesConfig} value={permissions} />
    }
  }

  return (
    <div className="d-flex flex-column">
      {getAllowedPolicies()}
      {isView ? (
        <div className="d-flex flex-row">
          <strong>Path:&nbsp;</strong>
          {prefix}
        </div>
      ) : (
        <CustomInput
          label="Allowed policies path:"
          value={prefix}
          placeholder="/img"
          onChanged={(event) => setPrefix(bucketId, event.target.value)}
        />
      )}
    </div>
  )
}

export default ObjectStorageUsersPermissionsPolicies
