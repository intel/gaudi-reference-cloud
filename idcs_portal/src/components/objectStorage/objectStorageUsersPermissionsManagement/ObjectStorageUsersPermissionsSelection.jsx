// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect } from 'react'
import ObjectStorageUsersPermissionsActions from './ObjectStorageUsersPermissionsActions'
import ObjectStorageUsersPermissionsPolicies from './ObjectStorageUsersPermissionsPolicies'
import ObjectStorageUsersPermissionsManagement from './ObjectStorageUsersPermissionsManagement'
import useBucketUsersPermissionsStore from '../../../store/bucketUsersPermissionsStore/BucketUsersPermissionsStore'
import CustomInput from '../../../utils/customInput/CustomInput'

const ObjectStorageUsersPermissionsSelection = (props) => {
  // props
  const isEdit = props?.isEdit || false
  const selectionType = useBucketUsersPermissionsStore((state) => state.selectionType)
  const setSelectionType = useBucketUsersPermissionsStore((state) => state.setSelectionType)
  const setBucketId = useBucketUsersPermissionsStore((state) => state.setBucketId)
  const setBucketsPermissions = useBucketUsersPermissionsStore((state) => state.setBucketsPermissions)

  useEffect(() => {
    setBucketSelectionType(selectionType)
  }, [selectionType])

  const setBucketSelectionType = (type) => {
    if (!isEdit) {
      setBucketsPermissions(null)
    }
    if (type === 'All') {
      setBucketId('All')
    }
    if (selectionType !== type) {
      setSelectionType(type)
    }
  }

  const applyPermissionConfig = {
    label: 'Apply permissions:',
    type: 'radio',
    customClass: 'text-nowrap',
    options: [
      {
        name: 'For all buckets',
        value: 'All',
        onChanged: () => {
          setBucketSelectionType('All')
        }
      },
      {
        name: 'Per bucket',
        value: 'bucket',
        onChanged: () => {
          setBucketSelectionType('bucket')
        }
      }
    ]
  }

  return (
    <div className="section px-0">
      <div className="d-flex flex-xs-column flex-md-row gap-s8">
        <CustomInput {...applyPermissionConfig} value={selectionType} />
        {selectionType === 'All' ? (
          <>
            <ObjectStorageUsersPermissionsActions bucketId={'All'} />
            <ObjectStorageUsersPermissionsPolicies bucketId={'All'} />
          </>
        ) : (
          <ObjectStorageUsersPermissionsManagement buckets={props.buckets} isEdit={isEdit} />
        )}
      </div>
    </div>
  )
}

export default ObjectStorageUsersPermissionsSelection
