// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import ObjectStorageUsersPermissionsBuckets from './ObjectStorageUsersPermissionsBuckets'
import ObjectStorageUsersPermissionsActions from './ObjectStorageUsersPermissionsActions'
import ObjectStorageUsersPermissionsPolicies from './ObjectStorageUsersPermissionsPolicies'
import { useState } from 'react'

const ObjectStorageUsersPermissionsManagement = (props) => {
  // props
  const isView = props.isView || false
  const isEdit = props.isEdit || false
  const buckets = props?.buckets || []

  const [bucket, setBucket] = useState(null)

  return (
    <>
      <ObjectStorageUsersPermissionsBuckets
        buckets={buckets}
        bucket={bucket}
        setBucket={setBucket}
        isView={isView}
        isEdit={isEdit}
      />
      {bucket && (
        <>
          <ObjectStorageUsersPermissionsActions bucketId={bucket} isView={isView} />
          <ObjectStorageUsersPermissionsPolicies bucketId={bucket} isView={isView} />
        </>
      )}
    </>
  )
}

export default ObjectStorageUsersPermissionsManagement
