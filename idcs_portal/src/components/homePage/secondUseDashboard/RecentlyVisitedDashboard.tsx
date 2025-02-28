// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Card, ListGroup } from 'react-bootstrap'
import { type VisitedSite } from '../../../store/appStore/AppStore'

interface RecentlyVisitedDashboardProps {
  visitedSites: VisitedSite[]
  onRedirectTo: (url: string) => void
}

const RecentlyVisitedDashboard: React.FC<RecentlyVisitedDashboardProps> = (props): JSX.Element => {
  const visitedSites = props.visitedSites
  const onRedirectTo = props.onRedirectTo
  return (
    <Card className="h-100">
      <Card.Body>
        <Card.Title as="h2" className="h6">
          Recently Visited
        </Card.Title>
        <ListGroup variant="flush" className="border-bottom">
          {visitedSites.map((site, idx) => (
            <ListGroup.Item
              key={idx}
              action
              intc-id={`link-dashboard-recently-visited-card-${site.title}`}
              data-wap_ref={`link-dashboard-recently-visited-card-${site.title}`}
              aria-label={`${site.title} link`}
              onClick={() => {
                onRedirectTo(site.path)
              }}
            >
              {site.title}
            </ListGroup.Item>
          ))}
        </ListGroup>
      </Card.Body>
    </Card>
  )
}
export default RecentlyVisitedDashboard
