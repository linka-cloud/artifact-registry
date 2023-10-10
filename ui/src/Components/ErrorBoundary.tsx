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

import { ErrorOutlineOutlined } from '@mui/icons-material'
import { Button, Collapse, Stack } from '@mui/material'
import React from 'react'
import { Code } from './Code'
import { TypingMessage } from './TypingMessage'

interface ErrorBoundaryProps {
  children: React.ReactNode;
}

interface ErrorBoundaryState {
  error: Error | null;
  errorInfo: React.ErrorInfo | null;
  details: boolean;
}

export class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props)
    this.state = { error: null, errorInfo: null, details: false }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    this.setState({
      error: error,
      errorInfo: errorInfo,
    })
  }

  toggleDetails = () => {
    this.setState({ details: !this.state.details })
  }

  render() {
    if (this.state.errorInfo) {
      return (
        <Stack alignItems='center' justifyContent='center' paddingTop={8}>
          <Stack>
            <ErrorOutlineOutlined
              sx={{
                fontSize: '24rem',
                padding: theme => theme.spacing(8),
                color: theme => theme.palette.text.secondary,
              }} />
          </Stack>
          <TypingMessage
            text={['Something went wrong.', 'Please open an issue on GitHub using the details below.']}
            speed={100}
            eraseDelay={3000}
            eraseSpeed={50}
            typingDelay={1000}
          />
          <Button onClick={this.toggleDetails}>{this.state.details ? 'Hide details' : 'Show details'}</Button>
          <Stack maxWidth='100%'>
            <Collapse in={this.state.details}>
              {this.state.error && <Code code={this.state.error.toString()} />}
              {this.state.errorInfo.componentStack && <Code code={this.state.errorInfo.componentStack} />}
            </Collapse>
          </Stack>
        </Stack>
      )
    }
    return this.props.children
  }
}
