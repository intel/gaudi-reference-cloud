// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import type React from 'react'

export const getCustomInputId = (label: string): string => {
  const labelForId = label.toString()
  return labelForId.replaceAll(' ', '').replaceAll('*', '').replaceAll('.', '').replace(':', '').trim()
}

export interface CustomInputOption {
  name: string
  dropSelect?: React.ReactElement | string | null
  defaultChecked?: boolean
  value?: number | string | null
  subTitleHtml?: React.ReactElement | string | null
  feedbackHtml?: React.ReactElement | string | null
  onChanged?: (event: CustomInputOnChangeEvent) => void
  disabled?: boolean
  hidden?: boolean
}

export interface CustomInputExtraButton {
  label: string
  buttonFunction: () => void
}

interface CustomInputDictionaryOption {
  key: CustomInputProps
  value: CustomInputProps
}

interface CustomInputGridCell {
  showField: boolean
  type: string
  canSelectRow: boolean
  value: any
  function: (value: any) => JSX.Element
}

type CustomInputGridOptions = Record<string, string | CustomInputGridCell>

interface CustomInputGridColumn {
  columnName: string
  targetColumn: string
  isSort?: boolean
  hideField?: boolean
  className?: string
  showOnBreakpoint?: boolean
  width?: string
}

export interface CustomInputOnChangeEvent {
  target: {
    value: string | number | null | undefined | boolean
  }
}

type CustomInputType =
  | 'multi-select'
  | 'dropdown'
  | 'textarea'
  | 'integer'
  | 'date'
  | 'checkbox'
  | 'radio'
  | 'radio-card'
  | 'grid'
  | 'text'
  | 'select'
  | 'switch'

type CustomInputSize = 'sm' | 'lg' | undefined

type FormControlElement = HTMLInputElement | HTMLTextAreaElement

export interface CustomInputProps {
  /***************************************************************/
  // Global Props
  /***************************************************************/

  /**
   * The autocomplete attribute for the input
   */
  autocomplete?: string
  /**
   * A string of classes to be applied to the input
   */
  customClass?: string
  /**
   * A button below the input
   */
  extraButton?: CustomInputExtraButton
  /**
   * A helper message for user
   */
  helperMessage?: React.ReactElement | string | null
  /**
   * If true the input is not visible
   */
  hidden?: boolean
  /**
   * If true the input label is not visible
   */
  hiddenLabel?: boolean
  /**
   * If true the input cannot be edited
   */
  isReadOnly?: boolean
  /**
   * If true the input is required
   */
  isRequired?: boolean
  /**
   * True if the user already interacted with the input
   */
  isTouched?: boolean
  /**
   * If true the input is valid
   * If false the input is invalid and error message is displayed
   */
  isValid?: boolean
  /**
   * The size of the input
   */
  fieldSize?: CustomInputSize
  /**
   * Custom Intc-Id for individual custom inputs (text, textArea)
   * Use just in case is needed otherwise let the default id
   * be created
   */
  intcId?: string
  /**
   * The maximum number of characters allowed
   */
  maxLength?: number
  /**
   * The maximum width of the Custom Input Component
   */
  maxWidth?: number
  /**
   * The maximum width of the Input
   */
  maxInputWidth?: number
  /**
   * The minimum number of characters allowed
   */
  minLength?: number
  /**
   * The Label for the input
   */
  label?: string
  /**
   * The Label for the input
   */
  subLabel?: React.ReactElement | string | null
  /**
   * A button link next to the label
   */
  labelButton?: CustomInputExtraButton
  /**
   * Triggered when input lost the focus
   */
  onBlur?: React.FocusEventHandler<FormControlElement>
  /**
   * Triggered when value is changed
   */
  onChanged?: (event: CustomInputOnChangeEvent) => void
  /**
   * Triggered when user enter key
   */
  onKeyDown?: React.KeyboardEventHandler<FormControlElement>
  /**
   * Collection of options for dropdown, multi-select, checkboxes or radios
   */
  options?: CustomInputOption[]
  /**
   * The placeholder for the input
   */
  placeholder?: string
  /**
   * A prepend element for input
   */
  prepend?: React.ReactElement | string | null
  /**
   * The type of input
   */
  type?: CustomInputType
  /**
   * The message displayed when input is on a non valid state
   */
  validationMessage?: string
  /**
   * The current value of the input
   */
  value?: any
  /**
   * The Minimum value for the input
   */
  min?: any

  /***************************************************************/
  // Dropdown Props
  /***************************************************************/

  /**
   * If true te dropdown multiple is painted in the screen as a list of checkboxes
   */
  borderlessDropdownMultiple?: boolean
  /**
   * A message shoed when multi-select has no options to pick
   */
  emptyOptionsMessage?: string
  /**
   * A method triggered after the selection changed,
   * receives the list of items selected
   */
  onChangeDropdownMultiple?: (values: any) => void
  /**
   * The select/unselect all button for the multi-select
   */
  selectAllButton?: CustomInputExtraButton
  /**
   * The refresh button for the multi-select
   */
  refreshButton?: CustomInputExtraButton

  /***************************************************************/
  // Dictionary Props
  /***************************************************************/

  /**
   * Collection of options for dictionary
   */
  dictionaryOptions?: CustomInputDictionaryOption[]
  /**
   * Call when input value changed
   */
  onChangeTagValue?: (e: CustomInputOnChangeEvent, type: 'key' | 'value', index: number) => void
  /**
   * Call when the add tag or delete tag button is pressed
   */
  onClickActionTag?: (index: number | null, action: 'Delete' | 'Add') => void

  /***************************************************************/
  // Textarea Props
  /***************************************************************/

  /**
   * The number of rows to calculate the height of text area
   */
  textAreaRows?: number

  /***************************************************************/
  // Radio group Props
  /***************************************************************/

  /**
   * If true the values are arrange in-line
   */
  radioGroupHorizontal?: boolean

  /***************************************************************/
  // Stepper Props
  /***************************************************************/

  /**
   * The maximum range for the Stepper CustomInput
   */
  maxRange?: number
  /**
   * The minimum range for the Stepper CustomInput
   */
  minRange?: number
  /**
   * Jump value for the Stepper CustomInput
   */
  step?: number

  /***************************************************************/
  // Multiselect Dropdown Props
  /***************************************************************/
  /**
   * If True, the filter option text will display
   */
  isFilter?: boolean

  /***************************************************************/
  // Validators Props
  /***************************************************************/

  /**
   * The meximum number allowed for Integer input
   */
  checkMaxValue?: number

  /***************************************************************/
  // Grid Props
  /***************************************************************/

  /**
   * Columns array for grid definition
   */
  columns?: CustomInputGridColumn[]
  /**
   * Name of the column that works as grid value
   */
  idField?: string
  /**
   * Flag to enable or disable single row selection on grid
   */
  singleSelection?: boolean
  /**
   * Array of values for the selected rows
   */
  selectedRecords?: any[]
  /**
   * Function that receives an array of values for the selected rows
   */
  setSelectedRecords?: (value: any[]) => void
  /**
   * Message to display when no data in grid
   */
  emptyGridMessage?: any
  /**
   * Breakpoint were only flagged columns are displayed
   */
  gridBreakpoint?: string
  /**
   * Content of the grid input row
   */
  gridOptions?: CustomInputGridOptions[]
  /**
   * Flag to show or hide pagination
   */
  hidePaginationControl?: boolean
}
