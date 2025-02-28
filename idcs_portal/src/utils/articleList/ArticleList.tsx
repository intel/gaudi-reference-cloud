// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { type ArticleItem } from '../../containers/navigation/Navigation.types'
import { Link } from 'react-router-dom'
import { ReactComponent as ExternalLink } from '../../assets/images/ExternalLink.svg'

interface ArticleListProps {
  items: ArticleItem[]
}

const ArticleList: React.FC<ArticleListProps> = ({ items }) => {
  return (
    <ul className="list-unstyled article-list">
      {items?.map((item, index) => (
        <li key={index} className="article-description">
          <Link
            className="h6 article-header"
            aria-label={`Learn ${item.alttitle}`}
            intc-id={`article-${item.alttitle.replace(' ', '_')}`}
            to={item.base_url}
            target={item.openInNewPage ? '_blank' : undefined}
          >
            {item.alttitle}
            <ExternalLink />
          </Link>
          <span className="small">{item.phrase}</span>
          <div className="article-subtitles">
            {item.show_urls &&
              item.urls.map((subtitleGroup) =>
                Object.entries(subtitleGroup).map(([subtitle, route], urlIndex: number) => (
                  <Link
                    key={urlIndex}
                    aria-label={`Learn ${subtitle}`}
                    intc-id={`article-subsection-${subtitle.replace(' ', '_')}`}
                    to={route}
                    target={item.openInNewPage ? '_blank' : undefined}
                  >
                    <span className="small text-decoration-underline">{subtitle}</span>
                  </Link>
                ))
              )}
          </div>
          <hr />
        </li>
      ))}
    </ul>
  )
}

export default ArticleList
