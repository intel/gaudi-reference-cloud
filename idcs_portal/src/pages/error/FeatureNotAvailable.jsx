// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import Button from 'react-bootstrap/Button'
import { getFeatureAvailableRegions } from '../../config/configurator'
import useAppStore from '../../store/appStore/AppStore'

const FeatureNotAvailable = ({ featureFlag = '' }) => {
  const changeRegion = useAppStore((state) => state.changeRegion)

  const [availableRegions, setAvailableRegions] = useState([])

  useEffect(() => {
    const availableRegions = getFeatureAvailableRegions(featureFlag)
    setAvailableRegions(availableRegions)
  }, [])

  return (
    <div className="section text-center align-items-center">
      <h1>Feature Unavailable</h1>
      <p>
        This feature is currently unavailable in your selected region.
        <br />
        Please choose a different region:
      </p>
      <div className="d-flex flex-column">
        {availableRegions.map((region, index) => (
          <Button
            variant="link"
            intc-id={`btn-switch-to-region-${region}`}
            key={index}
            aria-label={`Switch to region ${region}`}
            onClick={() => {
              changeRegion(region)
            }}
          >
            {region}
          </Button>
        ))}
      </div>
    </div>
  )
}

export default FeatureNotAvailable
