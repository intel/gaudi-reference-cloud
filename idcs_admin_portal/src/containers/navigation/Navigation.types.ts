// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { type IconType } from 'react-icons'

export interface IdcNavigation {
  name: string
  path: string
  externalPath?: string
  children?: IdcNavigation[]
  icon?: IconType | React.FunctionComponent<React.SVGProps<SVGSVGElement>>
  showInMenu?: boolean
  showInToolbar?: boolean
  dividerAtTop?: boolean
  showBadge?: boolean
}

export interface IdcBreadcrum {
  title: string
  path: string
  codePath: string
  hide: boolean
}

export interface ArticleItem {
  alttitle: string
  doctitle: string
  phrase: string
  base_url: string
  urls: object[]
  show_urls: boolean
  openInNewPage: boolean
  category: string[]
  keywords: string[]
  rating: number
}
