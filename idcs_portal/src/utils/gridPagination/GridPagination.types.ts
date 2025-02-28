// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import type React from 'react'
import { type CustomInputOption } from '../customInput/CustomInput.types'

export interface ColumnConfig {
  behaviorType: 'hyperLink' | 'icons' | 'buttons'
  behaviorFunction?: any | null
}

export interface ColumnDefinition {
  columnName: string
  targetColumn: string
  isSort?: boolean
  hideField?: boolean
  className?: string
  width?: string
}

export interface ButtonsCellDefinition {
  type?: 'icon' | 'button'
  name: string
  label: string
  variant?: string
  className?: string
}

export interface CellDefitition {
  type:
    | 'dropdown'
    | 'hyperlink'
    | 'checkbox'
    | 'buttons'
    | 'button'
    | 'currency'
    | 'date'
    | 'hyperlink-date'
    | 'function'
    | 'text'
  options?: CustomInputOption[]
  function?: any
  showField?: boolean
  value?: string | number | null
  format?: string
  toLocalTime?: boolean
  noHyperLinkValue?: string
  href?: string
  icon?: React.ReactElement
  selectableValues?: ButtonsCellDefinition[]
  colSpan?: number
  canSelectRow?: boolean
}
