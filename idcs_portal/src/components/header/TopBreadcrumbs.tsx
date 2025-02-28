// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Breadcrumb from 'react-bootstrap/Breadcrumb'
import { type IdcBreadcrum } from '../../containers/navigation/Navigation.types'
import { NavLink } from 'react-router-dom'
import { useMediaQuery } from 'react-responsive'
import { capitalizeString } from '../../utils/stringFormatHelper/StringFormatHelper'
import { BsArrowLeftShort } from 'react-icons/bs'

interface TopBreadcrumsProps {
  items: IdcBreadcrum[]
  currentPath: string
  shouldShowBreadcrum: boolean
}

const TopBreadcrums: React.FC<TopBreadcrumsProps> = ({ items, currentPath, shouldShowBreadcrum }) => {
  const isXsScreen = useMediaQuery({
    query: '(max-width: 575px)'
  })
  if (!shouldShowBreadcrum) {
    return null
  }
  if (items.length === 1) {
    return <Breadcrumb className="siteBreadcrum flex-row "></Breadcrumb>
  }
  return (
    <>
      <Breadcrumb className="siteBreadcrum flex-row align-self-start">
        {!isXsScreen ? (
          items.map((item: IdcBreadcrum, index) =>
            item.path === currentPath || index === items.length - 1 ? (
              <Breadcrumb.Item active key={index}>
                {capitalizeString(item.title)}
              </Breadcrumb.Item>
            ) : (
              <Breadcrumb.Item
                key={index}
                linkAs={NavLink}
                linkProps={{
                  to: item.path
                }}
              >
                {capitalizeString(item.title)}
              </Breadcrumb.Item>
            )
          )
        ) : (
          <Breadcrumb.Item
            aria-label={`Go Back to ${items[items.length - 2]?.title}`}
            linkAs={NavLink}
            linkProps={{
              to: items[items.length - 2]?.path
            }}
          >
            <BsArrowLeftShort />
            Go back
          </Breadcrumb.Item>
        )}
      </Breadcrumb>
    </>
  )
}

export default TopBreadcrums
