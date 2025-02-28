// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useState, useCallback } from 'react'

const useErrorBoundary = () => {
  // eslint-disable-next-line no-unused-vars
  const [_, setError] = useState()
  return useCallback(
    (e) => {
      if (!(e?.code === 'ERR_CANCELED')) {
        setError(() => {
          throw e
        })
      }
    },
    [setError]
  )
}

export default useErrorBoundary
