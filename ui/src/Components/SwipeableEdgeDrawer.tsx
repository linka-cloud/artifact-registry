// Copyright 2021 Linka Cloud  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { Global } from '@emotion/react'
import Box from '@mui/material/Box'
import { grey } from '@mui/material/colors'
import { styled } from '@mui/material/styles'
import SwipeableDrawer from '@mui/material/SwipeableDrawer'
import * as React from 'react'
import { useCallback, useEffect, useState } from 'react'

const drawerBleeding = 24
const pullerHeight = 6

const Root = styled('div')(({ theme }) => ({
  height: '100%',
  backgroundColor: theme.palette.mode === 'light' ? grey[100] : theme.palette.background.default,
}))

const StyledBox = styled(Box)(({ theme }) => ({
  backgroundColor: theme.palette.mode === 'light' ? '#fff' : grey[800],
}))

const Puller = styled(Box)(({ theme }) => ({
  width: 30,
  height: 6,
  backgroundColor: theme.palette.mode === 'light' ? grey[300] : grey[900],
  borderRadius: 3,
  position: 'absolute',
  top: 8,
  left: 'calc(50% - 15px)',
}))

export interface SwipeableDrawerProps extends React.PropsWithChildren<any> {
  onClose?: () => void
}

export const SwipeableEdgeDrawer = ({ children, onClose }: SwipeableDrawerProps) => {
  const [open, setOpen] = React.useState(false)

  const [height, setHeight] = useState(0)

  const heightChanged = useCallback((node: HTMLElement) => {
    console.log('height', node?.clientHeight, 'window', window.innerHeight)
    setHeight(node?.clientHeight > 40 ? node.clientHeight + drawerBleeding + pullerHeight : window.innerHeight / 2)
  }, [])

  const toggleDrawer = (newOpen: boolean) => () => {
    setOpen(newOpen)
  }

  const handleClose = () => {
    toggleDrawer(false)
    onClose?.()
  }
  useEffect(() => {
    setOpen(true)
  }, [])

  return (
    <Root>
      <Global
        styles={{
          '.MuiDrawer-root > .MuiPaper-root': {
            height: `calc(${height}px - ${drawerBleeding}px)`,
            overflow: 'visible',
          },
        }}
      />
      <SwipeableDrawer
        anchor='bottom'
        open={open}
        onClose={handleClose}
        onOpen={toggleDrawer(true)}
        swipeAreaWidth={drawerBleeding}
        disableSwipeToOpen={true}
        ModalProps={{
          keepMounted: false,
        }}
      >
        <StyledBox
          sx={{
            position: 'absolute',
            top: -drawerBleeding,
            borderTopLeftRadius: 8,
            borderTopRightRadius: 8,
            visibility: 'visible',
            right: 0,
            left: 0,
          }}
        >
          <Puller />
          <Box sx={{ height: drawerBleeding, paddingTop: `${pullerHeight + drawerBleeding}px` }} onClick={handleClose}>
            <Box ref={heightChanged}>{children}</Box>
          </Box>
        </StyledBox>
      </SwipeableDrawer>
    </Root>
  )
}
