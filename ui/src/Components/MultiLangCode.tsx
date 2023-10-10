// Copyright 2023 Linka Cloud  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { Tab as MuiTab, Tabs as MuiTabs } from '@mui/material'
import { styled } from '@mui/material/styles'
import React from 'react'
import { usePersistedState } from '../hooks'
import { Code, CodeProps } from './Code'

const Tabs = styled(MuiTabs)(({ theme }) => ({
  minHeight: 0,
  // background: 'black',
  '& .MuiTabs-indicator': {
    display: 'none'
  },
  '& .MuiTabs-flexContainer': {
    justifyContent: 'flex-end'
  }
}))

const Tab = styled(MuiTab)(({ theme }) => ({
  textTransform: 'none',
  // color: 'white',
  padding: `0 ${theme.spacing(2)}`,
  minWidth: 0,
  minHeight: 32,
  [theme.breakpoints.up('sm')]: {
    minWidth: 0,
    minHeight: 32,
  },
}))

export interface MultiLangCodeItemProps extends CodeProps {
  label: string
}

export const MultiLangCodeItem = (props: MultiLangCodeItemProps) => <Code {...props} />

export interface MultiLangCodeProps {
  key: string
  children: React.ReactElement<MultiLangCodeItemProps>| React.ReactElement<MultiLangCodeItemProps>[]
}

export const MultiLangCode = ({ key, children }: MultiLangCodeProps) => {
  const [value, setValue] = usePersistedState(0, 'MultiLangCode-' + key)
  const handleChange = (_: React.SyntheticEvent, newValue: number) => {
    setValue(newValue)
  }
  return (
    <>
      <Tabs
        value={value}
        onChange={handleChange}
      >
        { Array.isArray(children) ? children.map((e, i) => <Tab key={i} label={e.props.label} />) : <Tab label={children.props.label} />}
      </Tabs>
      <Code sx={{marginTop: 0}} {...(Array.isArray(children) ? children[value].props : children.props)} />
    </>
  )
}
