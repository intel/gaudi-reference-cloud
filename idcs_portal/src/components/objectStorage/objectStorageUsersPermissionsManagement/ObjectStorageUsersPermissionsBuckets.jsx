// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import useBucketUsersPermissionsStore from '../../../store/bucketUsersPermissionsStore/BucketUsersPermissionsStore'
import { useEffect } from 'react'
import CustomInput from '../../../utils/customInput/CustomInput'

const ObjectStorageUsersPermissionsBuckets = (props) => {
  // props
  const isView = props.isView || false
  const isEdit = props.isEdit || false

  const setBucket = props.setBucket
  const bucket = props.bucket
  const buckets = props.buckets

  const setBucketId = useBucketUsersPermissionsStore((state) => state.setBucketId)

  const selectBucket = (bucketName) => {
    setBucketId(bucketName)
    setBucket(bucketName)
  }

  useEffect(() => {
    if (buckets.length > 0 && (isView || isEdit)) {
      if (bucket) {
        selectBucket(bucket)
      } else {
        selectBucket(buckets[0].name)
      }
    }
  }, [bucket, buckets])

  const getSelectBuckets = () => {
    const selectBucketConfig = {
      label: 'Buckets:',
      type: 'radio',
      options: []
    }

    for (const index in buckets) {
      const bucket = { ...buckets[index] }
      selectBucketConfig.options.push({
        name: bucket.name,
        value: bucket.name,
        onChanged: () => {
          selectBucket(bucket.name)
        }
      })
    }
    return <CustomInput {...selectBucketConfig} value={bucket} />
  }

  return <>{getSelectBuckets()}</>
}

export default ObjectStorageUsersPermissionsBuckets
