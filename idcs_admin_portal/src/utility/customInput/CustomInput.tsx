// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

/* eslint-disable */
import React from 'react'
import CustomInputDropdown from './partials/CustomInputDropdown'
import CustomInputSelect from './partials/CustomInputSelect'
import CustomInputTextArea from './partials/CustomInputTextArea'
import CustomInputText from './partials/CustomInputText'
import CustomInputInteger from './partials/CustomInputInteger'
import CustomInputDate from './partials/CustomInputDate'
import CustomInputMultiselect from './partials/CustomInputMultiselect'
import CustomInputCheckbox from './partials/CustomInputCheckbox'
import CustomInputRadioGroup from './partials/CustomInputRadioGroup'
import { CustomInputProps } from './CustomInput.types'
import CustomInputDictionary from './partials/CustomInputDictionary'
import CustomInputRange from './partials/CustomInputRange'
import CustomInputSwitch from './partials/CustomInputSwitch'
import CustomInputRadioCard from './partials/CustomInputRadioCard'
import CustomInputGrid from './partials/CustomInputGrid'
import CustomInputDropdownMultiselect from './partials/CustomInputDropdownMultiselect'

const CustomInput: React.FC<CustomInputProps> = ({
  /***************************************************************/
  // Global Props
  /***************************************************************/
  autocomplete,
  customClass,
  extraButton,
  helperMessage,
  hidden,
  hiddenLabel,
  isReadOnly,
  isRequired,
  isTouched,
  isValid,
  fieldSize,
  maxLength,
  maxWidth,
  maxInputWidth,
  minLength,
  label = '',
  subLabel,
  labelButton,
  onBlur,
  onChanged,
  onKeyDown,
  options = [],
  placeholder,
  prepend,
  type = 'text',
  validationMessage,
  value,
  min,
  /***************************************************************/
  // Dropdown Props
  /***************************************************************/
  borderlessDropdownMultiple,
  emptyOptionsMessage,
  onChangeDropdownMultiple,
  selectAllButton,
  refreshButton,
  /***************************************************************/
  // Dictionary Props
  /***************************************************************/
  dictionaryOptions,
  onChangeTagValue,
  onClickActionTag,
  /***************************************************************/
  // Radio group Props
  /***************************************************************/
  radioGroupHorizontal,
  /***************************************************************/
  // Textarea Props
  /***************************************************************/
  textAreaRows,
  /***************************************************************/
  // Stepper Props
  /***************************************************************/
  maxRange,
  minRange,
  step,
  /***************************************************************/
  // Validators Props
  /***************************************************************/
  checkMaxValue,
  /***************************************************************/
  // Grid Props
  /***************************************************************/
  idField,
  columns,
  singleSelection,
  selectedRecords,
  setSelectedRecords,
  emptyGridMessage,
  gridBreakpoint,
  gridOptions,
  hidePaginationControl
}) => {
  if (hidden) {
    return <></>
  }
  // props functions
  let inputResponse = null

  switch (type.toUpperCase()) {
    case 'DROPDOWN':
      inputResponse = (
        <CustomInputDropdown
          label={label}
          labelButton={labelButton}
          maxWidth={maxWidth}
          maxInputWidth={maxInputWidth}
          options={options}
          isReadOnly={isReadOnly}
          hiddenLabel={hiddenLabel}
          value={value}
          placeholder={placeholder}
          onChanged={onChanged}
          extraButton={extraButton}
          validationMessage={validationMessage}
          isValid={isValid}
          isTouched={isTouched}
          helperMessage={helperMessage}
        />
      )
      break
    case 'DICTIONARY':
      inputResponse = (
        <CustomInputDictionary
          dictionaryOptions={dictionaryOptions}
          onChangeTagValue={onChangeTagValue}
          label={label}
          subLabel={subLabel}
          labelButton={labelButton}
          hiddenLabel={hiddenLabel}
          onClickActionTag={onClickActionTag}
          maxLength={maxLength}
          isValid={isValid}
          isTouched={isTouched}
          validationMessage={validationMessage}
          helperMessage={helperMessage}
        />
      )
      break
    case 'TEXTAREA':
      inputResponse = (
        <CustomInputTextArea
          label={label}
          labelButton={labelButton}
          fieldSize={fieldSize}
          isReadOnly={isReadOnly}
          hiddenLabel={hiddenLabel}
          value={value}
          placeholder={placeholder}
          onChanged={onChanged}
          onBlur={onBlur}
          extraButton={extraButton}
          helperMessage={helperMessage}
          maxLength={maxLength}
          textAreaRows={textAreaRows}
          validationMessage={validationMessage}
          isValid={isValid}
          isTouched={isTouched}
          customClass={customClass}
        />
      )
      break
    case 'MULTI-SELECT':
      inputResponse = (
        <CustomInputMultiselect
          label={label}
          options={options}
          labelButton={labelButton}
          emptyOptionsMessage={emptyOptionsMessage}
          fieldSize={fieldSize}
          isReadOnly={isReadOnly}
          hiddenLabel={hiddenLabel}
          value={value}
          onChangeDropdownMultiple={onChangeDropdownMultiple}
          onBlur={onBlur}
          extraButton={extraButton}
          refreshButton={refreshButton}
          selectAllButton={selectAllButton}
          helperMessage={helperMessage}
          validationMessage={validationMessage}
          isValid={isValid}
          isTouched={isTouched}
          borderlessDropdownMultiple={borderlessDropdownMultiple}
          customClass={customClass}
        />
      )
      break
    case 'SELECT':
      inputResponse = (
        <CustomInputSelect
          label={label}
          labelButton={labelButton}
          fieldSize={fieldSize}
          options={options}
          isReadOnly={isReadOnly}
          hiddenLabel={hiddenLabel}
          value={value}
          autocomplete={autocomplete}
          placeholder={placeholder}
          onChanged={onChanged}
          extraButton={extraButton}
        />
      )
      break
    case 'INTEGER':
      inputResponse = (
        <CustomInputInteger
          label={label}
          labelButton={labelButton}
          fieldSize={fieldSize}
          isReadOnly={isReadOnly}
          hiddenLabel={hiddenLabel}
          autocomplete={autocomplete}
          value={value}
          placeholder={placeholder}
          onChanged={onChanged}
          onBlur={onBlur}
          extraButton={extraButton}
          helperMessage={helperMessage}
          maxLength={maxLength}
          checkMaxValue={checkMaxValue}
          validationMessage={validationMessage}
          isValid={isValid}
          isTouched={isTouched}
          customClass={customClass}
        />
      )
      break
    case 'DATE':
      inputResponse = (
        <CustomInputDate
          label={label}
          labelButton={labelButton}
          fieldSize={fieldSize}
          isReadOnly={isReadOnly}
          hiddenLabel={hiddenLabel}
          autocomplete={autocomplete}
          value={value}
          placeholder={placeholder}
          onChanged={onChanged}
          onBlur={onBlur}
          extraButton={extraButton}
          helperMessage={helperMessage}
          validationMessage={validationMessage}
          isValid={isValid}
          isTouched={isTouched}
          min={min}
        />
      )
      break
    case 'RADIO':
      inputResponse = (
        <CustomInputRadioGroup
          label={label}
          labelButton={labelButton}
          customClass={customClass}
          options={options}
          fieldSize={fieldSize}
          isReadOnly={isReadOnly}
          hiddenLabel={hiddenLabel}
          value={value}
          onChanged={onChanged}
          onBlur={onBlur}
          extraButton={extraButton}
          helperMessage={helperMessage}
          validationMessage={validationMessage}
          radioGroupHorizontal={radioGroupHorizontal}
          isValid={isValid}
          isTouched={isTouched}
        />
      )
      break
    case 'RADIO-CARD':
      inputResponse = (
        <CustomInputRadioCard
          label={label}
          labelButton={labelButton}
          options={options}
          hiddenLabel={hiddenLabel}
          value={value}
          onChanged={onChanged}
          helperMessage={helperMessage}
          validationMessage={validationMessage}
          extraButton={extraButton}
          isValid={isValid}
          isTouched={isTouched}
        />
      )
      break
    case 'GRID':
      inputResponse = (
        <CustomInputGrid
          idField={idField}
          label={label}
          labelButton={labelButton}
          maxWidth={maxWidth}
          hiddenLabel={hiddenLabel}
          singleSelection={singleSelection}
          gridOptions={gridOptions}
          columns={columns}
          selectedRecords={selectedRecords}
          setSelectedRecords={setSelectedRecords}
          emptyGridMessage={emptyGridMessage}
          gridBreakpoint={gridBreakpoint}
          hidePaginationControl={hidePaginationControl}
        />
      )
      break
    case 'CHECKBOX':
      inputResponse = (
        <CustomInputCheckbox
          label={label}
          labelButton={labelButton}
          options={options}
          fieldSize={fieldSize}
          isReadOnly={isReadOnly}
          hiddenLabel={hiddenLabel}
          value={value}
          onChanged={onChanged}
          onBlur={onBlur}
          extraButton={extraButton}
          helperMessage={helperMessage}
          validationMessage={validationMessage}
          isValid={isValid}
          isTouched={isTouched}
          maxWidth={maxWidth}
        />
      )
      break
    case 'RANGE':
      inputResponse = (
        <CustomInputRange
          label={label}
          labelButton={labelButton}
          fieldSize={fieldSize}
          isReadOnly={isReadOnly}
          hiddenLabel={hiddenLabel}
          minRange={minRange}
          maxRange={maxRange}
          value={value}
          step={step}
          onChanged={onChanged}
          onBlur={onBlur}
          extraButton={extraButton}
          helperMessage={helperMessage}
          validationMessage={validationMessage}
          isValid={isValid}
          isTouched={isTouched}
        />
      )
      break
    case 'SWITCH':
      inputResponse = (
        <CustomInputSwitch
          label={label}
          labelButton={labelButton}
          options={options}
          fieldSize={fieldSize}
          isReadOnly={isReadOnly}
          hiddenLabel={hiddenLabel}
          value={value}
          onChanged={onChanged}
          onBlur={onBlur}
          extraButton={extraButton}
          helperMessage={helperMessage}
          validationMessage={validationMessage}
          isValid={isValid}
          isTouched={isTouched}
        />
      )
      break
    case 'MULTI-SELECT-DROPDOWN':
      inputResponse = (
        <CustomInputDropdownMultiselect
          label={label}
          options={options}
          labelButton={labelButton}
          emptyOptionsMessage={emptyOptionsMessage}
          fieldSize={fieldSize}
          isReadOnly={isReadOnly}
          hiddenLabel={hiddenLabel}
          value={value}
          onChangeDropdownMultiple={onChangeDropdownMultiple}
          maxWidth={maxWidth}
          extraButton={extraButton}
          refreshButton={refreshButton}
          placeholder={placeholder}
          onChanged={onChanged}
          validationMessage={validationMessage}
          isValid={isValid}
          isTouched={isTouched}
          helperMessage={helperMessage}
          customClass={customClass}
          selectAllButton={selectAllButton}
        />
      )
      break
    default:
      inputResponse = (
        <CustomInputText
          label={label}
          labelButton={labelButton}
          maxWidth={maxWidth}
          fieldSize={fieldSize}
          isReadOnly={isReadOnly}
          hiddenLabel={hiddenLabel}
          value={value}
          placeholder={placeholder}
          onChanged={onChanged}
          onBlur={onBlur}
          onKeyDown={onKeyDown}
          extraButton={extraButton}
          helperMessage={helperMessage}
          prepend={prepend}
          maxLength={maxLength}
          minLength={minLength}
          validationMessage={validationMessage}
          isValid={isValid}
          isTouched={isTouched}
          customClass={customClass}
        />
      )
      break
  }

  return <>{inputResponse}</>
}

export default CustomInput
