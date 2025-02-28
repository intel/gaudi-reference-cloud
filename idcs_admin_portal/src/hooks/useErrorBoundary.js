import { useState, useCallback } from 'react'

const useErrorBoundary = () => {
  // eslint-disable-next-line no-unused-vars
  const [_, setError] = useState()
  return useCallback(
    e => {
      setError(() => {
        throw e
      })
    },
    [setError]
  )
}

export default useErrorBoundary
