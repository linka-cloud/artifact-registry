import { Clear } from '@mui/icons-material'

import { Button, IconButton, TextFieldProps as MuiTextFieldProps } from '@mui/material'
import { Field, FieldArray, FieldProps, getIn } from 'formik'

import {
  Autocomplete as FMuiAutocomplete,
  AutocompleteProps as FMuiAutocompleteProps,
  Checkbox as FMuiCheckbox,
  CheckboxProps as FMuiCheckboxProps,
  CheckboxWithLabel as FMuiCheckboxWithLabel,
  CheckboxWithLabelProps as FMuiCheckboxWithLabelProps,
  InputBase as FMuiInputBase,
  InputBaseProps as FMuiInputBaseProps,
  RadioGroup as FMuiRadioGroup,
  RadioGroupProps as FMuiRadioGroupProps,
  Select as FMuiSelect,
  SelectProps as FMuiSelectProps,
  SimpleFileUpload as FMuiSimpleFileUpload,
  SimpleFileUploadProps as FMuiSimpleFileUploadProps,
  Switch as FMuiSwitch,
  SwitchProps as FMuiSwitchProps,
  TextField as FMuiTextField,
  TextFieldProps as FMuiTextFieldProps,
  ToggleButtonGroup as FMuiToggleButtonGroup,
  ToggleButtonGroupProps as FMuiToggleButtonGroupProps,
} from 'formik-mui'
import { FieldAttributes } from 'formik/dist/Field'

import React from 'react'

export const FAutocomplete = <T,
  U,
  Multiple extends boolean | undefined,
  DisableClearable extends boolean | undefined,
  FreeSolo extends boolean | undefined,
>(
  props: Omit<FMuiAutocompleteProps<T, Multiple, DisableClearable, FreeSolo>, keyof FieldProps> & FieldAttributes<U>,
) => <Field component={FMuiAutocomplete} {...props} />
export const FTextField = <T extends unknown>(props: Omit<FMuiTextFieldProps, keyof FieldProps> & FieldAttributes<T>) => (
  <Field component={FMuiTextField} {...props} />
)
export const FCheckbox = <T extends unknown>(props: Omit<FMuiCheckboxProps, keyof FieldProps> & FieldAttributes<T>) => (
  <Field component={FMuiCheckbox} {...props} />
)
export const FCheckboxWithLabel = <T extends unknown>(props: Omit<FMuiCheckboxWithLabelProps, keyof FieldProps> & FieldAttributes<T>) => (
  <Field component={FMuiCheckboxWithLabel} {...props} />
)
export const FInputBase = <T extends unknown>(props: Omit<FMuiInputBaseProps, keyof FieldProps> & FieldAttributes<T>) => (
  <Field component={FMuiInputBase} {...props} />
)
export const FRadioGroup = <T extends unknown>(props: Omit<FMuiRadioGroupProps, keyof FieldProps> & FieldAttributes<T>) => (
  <Field component={FMuiRadioGroup} {...props} />
)
export const FSelect = <T extends unknown>(props: Omit<FMuiSelectProps, keyof FieldProps> & FieldAttributes<T>) => (
  <Field component={FMuiSelect} {...props} />
)
export const FSimpleFileUpload = <T extends unknown>(props: Omit<FMuiSimpleFileUploadProps, keyof FieldProps> & FieldAttributes<T>) => (
  <Field component={FMuiSimpleFileUpload} {...props} />
)
export const FSwitch = <T extends unknown>(props: Omit<FMuiSwitchProps, keyof FieldProps> & FieldAttributes<T>) => (
  <Field component={FMuiSwitch} {...props} />
)
export const FToggleButtonGroup = <T extends unknown>(props: Omit<FMuiToggleButtonGroupProps, keyof FieldProps> & FieldAttributes<T>) => (
  <Field component={FMuiToggleButtonGroup} {...props} />
)

interface FFieldArrayElementProps<Values> {
  id: string;
  key: any;
  name: string;
  label?: string;
  onRemove: () => void;
  error: any;
  disabled: boolean;
  onBlur: any;
  onChange: any;
}

export interface AddElementProps {
  onAdd: () => void;
  disabled: boolean;
  text?: string;
}

interface FFieldArrayProps<Values> {
  name: string;
  label: string;
  Container: React.ElementType<{ children: React.ReactNode }>;
  Element: React.ElementType<FFieldArrayElementProps<Values>>;
  createNew: () => Partial<Values>;
  Add?: React.ElementType<AddElementProps>;
  addButtonText?: string;
}

export const FFieldArray = <T extends unknown>({
                                                 name,
                                                 label,
                                                 createNew,
                                                 Container,
                                                 Element,
                                                 Add,
                                                 addButtonText,
                                                 ...props
                                               }: FFieldArrayProps<T>) => (
  //@ts-ignore
  <FieldArray
    name={name.toString()}
    render={({ remove, replace, push, form: { values, errors, isSubmitting, touched, handleBlur } }) => (
      <>
        {
          //@ts-ignore
          getIn(values, name)?.map((v: unknown, i: number) => (
            <Container key={i}>
              <Element
                {...props}
                key={i}
                //@ts-ignore
                id={`${name}.${i}`}
                label={label}
                //@ts-ignore
                name={`${name}.${i}`}
                error={!!(getIn(errors, name) as any)?.[i] && (getIn(touched, name) as any)?.[i]}
                onRemove={() => remove(i)}
                disabled={isSubmitting}
                onChange={(e: any) => replace(i, e.target.value)}
                onBlur={handleBlur}
              />
            </Container>
          ))
        }
        {/*@ts-ignore*/}
        <Add disabled={isSubmitting} onAdd={() => push(createNew())} text={addButtonText} />
      </>
    )}
  />
)

export interface FTextFieldArrayProps extends Omit<FFieldArrayProps<string>, 'Container' | 'Element' | 'createNew'> {
  textFieldProps?: Omit<MuiTextFieldProps, 'name' | 'value' | 'error'>;
  deleteIcon?: React.ReactNode;
  Container?: React.ElementType<{ children: React.ReactNode }>;
  Element?: React.ElementType<FFieldArrayElementProps<string>>;
  createNew?: () => Partial<string>;
  addButtonText?: string;
}

export const FTextFieldArray = ({
                                  deleteIcon,
                                  createNew,
                                  Element,
                                  Add,
                                  Container,
                                  ...props
                                }: FTextFieldArrayProps) => (
  <FFieldArray {...props} createNew={createNew || (() => '')} Element={Element || FTextFieldArrayElement}
               Add={Add || AddElement}
               Container={Container || React.Fragment} />
)

interface FTextFieldArrayElementProps extends FFieldArrayElementProps<string> {
  textFieldProps?: Omit<MuiTextFieldProps, 'name' | 'value' | 'error'>;
  deleteIcon?: React.ReactNode;
}

export const FTextFieldArrayElement = ({
                                         onRemove,
                                         textFieldProps,
                                         deleteIcon,
                                         ...props
                                       }: FTextFieldArrayElementProps) => (
  //@ts-ignore
  <FTextField
    {...props}
    {...textFieldProps}
    InputProps={{
      ...props,
      endAdornment: <IconButton onClick={onRemove}>{deleteIcon ? deleteIcon : <Clear />}</IconButton>,
    }}
  />
)

const AddElement = ({ onAdd, disabled, text }: AddElementProps) => (
  <Button disabled={disabled} onClick={onAdd}>
    {text || 'Add'}
  </Button>
)
