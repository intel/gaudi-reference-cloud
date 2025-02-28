// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Card from 'react-bootstrap/Card'
import ButtonGroup from 'react-bootstrap/ButtonGroup'
import { BsBookmarkStarFill } from 'react-icons/bs'
import { NavLink } from 'react-router-dom'
import './FeaturedProductCard.scss'

interface FeaturedProductCardProps {
  topImage: React.ReactElement | null
  title: string
  description: string
  learnMoreHref?: string
}

const FeaturedProductCard: React.FC<FeaturedProductCardProps> = (props): JSX.Element => {
  const { topImage, title, description, learnMoreHref } = props

  return (
    <Card className="bg-surface-third-party-product featured-product-card p-s8">
      <div className="featured-product-card-body px-s10 gap-s6">
        {topImage && <div className="featured-product-card-itemgap-s4">{topImage}</div>}
        <div className="featured-product-card-item flex-column">
          <span className="display-6">{title}</span>
          <span className="lead">{description}</span>
        </div>
        <ButtonGroup className="featured-product-card-item align-items-center flex-row gap-s6">
          {learnMoreHref && (
            <NavLink
              intc-id={`learnmore-${title}`}
              to={learnMoreHref}
              onClick={(e) => {
                e.stopPropagation()
              }}
              className="btn btn-primary"
              aria-label={`Learn more about ${title}`}
            >
              Learn more
            </NavLink>
          )}
          <span className="featureTag gap-s4">
            <BsBookmarkStarFill />
            Featured Product
          </span>
        </ButtonGroup>
      </div>
    </Card>
  )
}

export default FeaturedProductCard
