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

import { styled } from '@mui/material/styles'
import ReactTypingEffect from 'react-typing-effect'

export const TypingMessage = styled(ReactTypingEffect)(({ theme }) => ({
  width: '100%',

  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',

  fontFamily: 'Courier New,Courier,Lucida Sans Typewriter,Lucida Typewriter,monospace',
  backgroundColor: theme.palette.background.default,
  color: theme.palette.text.secondary,
  fontSize: '3rem',
  marginBottom: theme.spacing(8),
  '& > span': {
    backgroundColor: theme.palette.background.default,
  },
}))
