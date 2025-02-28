// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Card from 'react-bootstrap/Card'
import { Badge, Button } from 'react-bootstrap'
import { type GetStartedCard } from '../../containers/homePage/Home.types'

interface DashboardInfoCardProps {
  cardInfo: GetStartedCard
  className?: string
  onRedirectTo: (value: string) => void
}

const DashboardInfoCard: React.FC<DashboardInfoCardProps> = (props): JSX.Element => {
  const cardInfo = props.cardInfo
  const className = props.className
  const onRedirectTo = props.onRedirectTo

  return (
    <Card className={`pb-s5 h-100 ${className ?? ''}`}>
      <Card.Img
        variant="top"
        src={cardInfo.imgSrc}
        srcSet={cardInfo.imgSrcSet}
        alt={`${cardInfo.title} image`}
        className="img-fluid ratio ratio-21x9"
      />
      <Card.Body className="flex-grow-0 h-100">
        <Card.Title as="h2" className="h5 d-flex flex-row gap-s4">
          {cardInfo.title}
          {cardInfo?.badge && (
            <Badge bg="primary" className="mb-0">
              {cardInfo?.badge}
            </Badge>
          )}
        </Card.Title>
        <Card.Subtitle className="text-muted">{cardInfo.subTitle}</Card.Subtitle>
        <Card.Text as="span" className="mb-s4">
          {cardInfo.homePageText}
        </Card.Text>
        <Button
          className="mt-auto"
          variant="outline-primary"
          intc-id={`btn-dashboard-info-card-${cardInfo.title}`}
          data-wap_ref={`btn-dashboard-info-card-${cardInfo.title}`}
          aria-label={`Get started to ${cardInfo.title} tab button`}
          onClick={() => {
            onRedirectTo(cardInfo.redirectTo)
          }}
        >
          Get Started
        </Button>
      </Card.Body>
    </Card>
  )
}

export default DashboardInfoCard
