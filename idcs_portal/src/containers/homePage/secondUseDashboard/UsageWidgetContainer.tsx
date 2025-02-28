// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import UsageWidget from '../../../components/homePage/secondUseDashboard/UsageWidget'
import useUsagesReport from '../../../store/billingStore/UsagesStore'
import useCloudCreditsStore from '../../../store/billingStore/CloudCreditsStore'

interface UsageWidgetContainerProps {
  showCloudCredits: boolean
}

const UsageWidgetContainer: React.FC<UsageWidgetContainerProps> = (props): JSX.Element => {
  const usageLoading = useUsagesReport((state) => state.loading)
  const creditsLoading = useCloudCreditsStore((state) => state.loading)
  const setUsage = useUsagesReport((state) => state.setUsage)
  const totalAmount = useUsagesReport((state) => state.totalAmount)
  const remainingCredits = useCloudCreditsStore((state) => state.remainingCredits)
  const setCloudCredits = useCloudCreditsStore((state) => state.setCloudCredits)
  const [usageError, setUsageError] = useState(false)
  const [creditsError, setCreditsError] = useState(false)

  const fetchUsage = async (): Promise<void> => {
    try {
      await setUsage()
      setUsageError(false)
    } catch {
      setUsageError(true)
    }
  }

  const fetchCredits = async (): Promise<void> => {
    try {
      await setCloudCredits()
      setCreditsError(false)
    } catch {
      setCreditsError(true)
    }
  }

  // Hooks
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      const promiseArray = []
      promiseArray.push(fetchUsage())
      promiseArray.push(fetchCredits())
      await Promise.allSettled(promiseArray)
    }
    void fetch()
  }, [])

  return (
    <UsageWidget
      usageLoading={usageLoading}
      creditsLoading={creditsLoading}
      totalAmount={totalAmount}
      remainingCredits={remainingCredits}
      showCloudCredits={props.showCloudCredits}
      usageError={usageError}
      creditsError={creditsError}
    />
  )
}

export default UsageWidgetContainer
