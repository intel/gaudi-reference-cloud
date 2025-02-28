// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useRef } from 'react'
import Card from 'react-bootstrap/Card'
import ButtonGroup from 'react-bootstrap/ButtonGroup'
import { NavLink } from 'react-router-dom'
import { BsArrowRightShort } from 'react-icons/bs'
import './ProductCard.scss'

interface ProductCardProps {
  'intc-id'?: string
  title: string
  description: string
  actionLabel?: string
  imageSource?: string
  onClick?: React.MouseEventHandler<HTMLAnchorElement>
  href?: string
  learnMoreHref?: string
  pricing?: string
}

const ProductCard: React.FC<ProductCardProps> = (props): JSX.Element => {
  const { title, description, actionLabel, onClick, href, learnMoreHref, imageSource, pricing } = props
  const anchorRef = useRef<HTMLAnchorElement>(null)

  return (
    <Card
      className={`bg-surface-gradient product-card ${learnMoreHref ? 'product-card-noclick' : ''}`}
      onClick={(e) => {
        e.preventDefault()
        if (anchorRef.current) {
          anchorRef.current.click()
        }
      }}
    >
      <Card.Body>
        <div className="d-flex flex-row w-100 gap-s6">
          {imageSource && <img src={imageSource} alt="..." className="product-icon" />}
          <div className="d-flex flex-column gap-s4">
            <span intc-id={`title-${title}`} className="fw-semibold">
              {title}
            </span>
            <span>{description}</span>
          </div>
          <div className="d-flex ms-auto mb-auto">
            <NavLink
              ref={anchorRef}
              intc-id={learnMoreHref ? `learnmore-${title}` : props['intc-id']}
              to={learnMoreHref ?? href ?? '#'}
              onClick={
                learnMoreHref
                  ? undefined
                  : (e) => {
                      e.stopPropagation()
                      if (onClick) {
                        onClick(e)
                      }
                    }
              }
              className="btn btn-link"
              aria-label={learnMoreHref ? `Learn more about ${title}` : `${actionLabel ?? 'Launch'} ${title}`}
            >
              {!learnMoreHref && (
                <>
                  <span className="launch-button-span">{actionLabel ?? 'Launch'}</span>
                  <BsArrowRightShort />
                </>
              )}
            </NavLink>
          </div>
        </div>
        <div className="d-flex w-100 pt-s6 mt-auto">
          {learnMoreHref && (
            <ButtonGroup>
              <NavLink
                intc-id={props['intc-id']}
                to={href ?? '#'}
                onClick={(e) => {
                  e.stopPropagation()
                  if (onClick) {
                    onClick(e)
                  }
                }}
                className="btn btn-outline-primary"
                aria-label={`${actionLabel ?? 'Launch'} ${title}`}
              >
                {actionLabel ?? 'Launch'}
              </NavLink>
            </ButtonGroup>
          )}
          {pricing ? <span className="ms-auto align-self-end">{pricing}</span> : null}
        </div>
      </Card.Body>
    </Card>
  )
}

export default ProductCard
