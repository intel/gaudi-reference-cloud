// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import useBucketUsersPermissionsStore from '../../../store/bucketUsersPermissionsStore/BucketUsersPermissionsStore'
import { objectStorageUsersPermissionActions } from '../../../utils/Enums'
import CustomInput from '../../../utils/customInput/CustomInput'

const ObjectStorageUsersPermissionsActions = (props) => {
  // props
  const isView = props.isView || false
  const bucketId = props.bucketId

  const bucketsPermissions = useBucketUsersPermissionsStore((state) => state.bucketsPermissions)
  const setActions = useBucketUsersPermissionsStore((state) => state.setActions)

  const selectedBucket = bucketsPermissions !== null ? bucketsPermissions[bucketId] : null
  const actions = selectedBucket?.actions ?? []

  const onSelectAllActions = () => {
    const objectKeys = Object.keys(objectStorageUsersPermissionActions)
    const selectedValues = actions
    const shouldDeselect = objectKeys.every((x) => selectedValues.includes(x))
    setActions(bucketId, shouldDeselect ? [] : objectKeys)
  }

  const clickMultipleSelectValues = (selectedOptions) => {
    setActions(bucketId, selectedOptions)
  }

  const getAllowedActions = () => {
    if (isView) {
      return (
        <div className="d-flex flex-column gap-s4">
          <label className="mb-0 form-label">Allowed actions:</label>
          {actions.map((item, index) => (
            <div key={index}>
              <span>{item}</span>
            </div>
          ))}
        </div>
      )
    } else {
      const allowedActionsConfig = {
        label: 'Allowed actions:',
        type: 'multi-select',
        customClass: 'text-nowrap',
        borderlessDropdownMultiple: true,
        onChangeDropdownMultiple: clickMultipleSelectValues,
        options: [],
        selectAllButton: {
          label: 'Select/Deselect All',
          buttonFunction: () => onSelectAllActions()
        }
      }
      for (const index in objectStorageUsersPermissionActions) {
        const name = objectStorageUsersPermissionActions[index]
        allowedActionsConfig.options.push({
          name,
          value: index
        })
      }
      return <CustomInput {...allowedActionsConfig} value={actions} />
    }
  }

  return <>{getAllowedActions()}</>
}

export default ObjectStorageUsersPermissionsActions
