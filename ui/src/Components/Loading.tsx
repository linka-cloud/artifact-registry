// Copyright 2022 Linka Cloud  All rights reserved.
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

import styled from '@emotion/styled'
import { Box, keyframes, useTheme } from '@mui/material'
import React from 'react'

interface Props {
  size?: string
  margin?: string
  background?: string
  duration?: string
  dots?: any
}

const defaultProps: Props = {
  dots: 3,
}

export const LoadingDots: React.FC<Props> = ({ size, margin, background, duration, dots }) => {
  const Wraper = styled.div`
    display: flex;
    justify-content: center;
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
  `

  const bounceLoading = keyframes`
    to {
      opacity: 0.3;
      transform: translate3d(0, -1.5rem, 0);
    }
  `

  const Dot = styled.div`
    width: ${size ? size : '1.5rem'};
    height: ${size ? size : '1.5rem'};
    margin: 0 ${margin ? margin : '1rem'};
    background: ${background ? background : 'rgb(202, 57, 57)'};
    border-radius: 50%;
    animation: ${duration ? duration : '0.8s'} ${bounceLoading} infinite alternate;

    &:nth-of-type(2n + 0) {
      animation-delay: 0.3s;
    }

    &:nth-of-type(3n + 0) {
      animation-delay: 0.6s;
    }
  `

  let dotList = []
  for (let i = 0; i < dots; i++) {
    dotList.push(i)
  }

  const dotRender = dotList.map((dot) => <Dot key={dot}></Dot>)

  return <Wraper>{dotRender}</Wraper>
}

LoadingDots.defaultProps = defaultProps

export const Loading = React.memo(() => {
  const theme = useTheme()
  return (
    <Box
      sx={{
        flex: 1,
      }}
    >
      <LoadingDots background={theme.palette.primary.main} size='1.2rem' />
    </Box>
  )
})
