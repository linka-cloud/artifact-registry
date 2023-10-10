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

import { ClearOutlined } from '@mui/icons-material'
import { createTheme, PaletteMode, PaletteOptions, ThemeOptions } from '@mui/material'
import { grey } from '@mui/material/colors'
import { LinkProps } from '@mui/material/Link'
import { PaletteColorOptions } from '@mui/material/styles/createPalette'
import React from 'react'
import { Link as RouterLink, LinkProps as RouterLinkProps } from 'react-router-dom'

const LinkBehavior = React.forwardRef<any, Omit<RouterLinkProps, 'to'> & {
  href: RouterLinkProps['to']
}>((props, ref) => {
  const { href, ...other } = props
  return <RouterLink data-testid='custom-link' ref={ref} to={href} {...other} />
})

export const configureTheme = (mode: PaletteMode) => {
  const palette = mode === 'dark' ? darkPalette : lightPalette
  return createTheme({
    ...makeTheme(palette),
    palette,
  })
}

const successPalette: PaletteColorOptions = {
  main: '#2e7d32',
}

const errorsPalette: PaletteColorOptions = {
  main: '#bb0a0a',
}

const lightPalette: PaletteOptions = {
  mode: 'light',
  primary: {
    main: '#006573',
    // main: '#2E586B',
  },
  background: {
    default: '#f6f6f6',
  },
  error: errorsPalette,
  success: successPalette,
}

const darkPalette: PaletteOptions = {
  mode: 'dark',
  primary: {
    main: '#006573',
  },
  background: {
    default: '#0e0e0e',
    // paper: '#121212'
    paper: '#191919',
  },
  error: errorsPalette,
  success: successPalette,
}

const makeTheme = (palette: PaletteOptions): ThemeOptions => {
  const theme = createTheme({
    palette,
    shape: {
      borderRadius: 4,
    },
    spacing: 8,
    typography: {
      fontFamily: [
        'Metropolis',
        '-apple-system',
        'BlinkMacSystemFont',
        'Segoe UI',
        'Roboto',
        'Oxygen',
        'Helvetica Neue',
        'sans-serif',
        // 'Quicksand',
        // 'sans-serif',
        // 'Arial',
        // 'sans-serif',
      ].join(','),
    },
  })
  return {
    ...theme,
    components: {
      MuiStack: {
        defaultProps: {
          spacing: theme.spacing(2),
        },
      },
      MuiPopover: {
        defaultProps: {
          elevation: 1,
        },
      },
      MuiAppBar: {
        defaultProps: {
          elevation: 0,
          position: 'fixed',
          variant: 'outlined',
        },
        styleOverrides: {
          root: {
            borderTop: 0,
            borderRight: 0,
            borderLeft: 0,
          },
        },
      },
      MuiDrawer: {
        defaultProps: {
          elevation: 0,
        },
      },
      MuiBottomNavigation: {
        styleOverrides: {
          root: {
            // borderRadius: '20% 20% 0% 0%'
            background: 'none',
          },
        },
      },
      MuiList: {
        styleOverrides: {
          padding: {
            padding: 0,
          },
        },
      },
      MuiListItem: {
        styleOverrides: {
          root: {
            '&.Mui-selected': {
              borderLeft: '2px',
              borderStyle: 'solid',
              boxSizing: 'content-box',
            },
          },
        },
      },
      MuiSkeleton: {
        defaultProps: {
          animation: 'wave',
        },
      },
      MuiLink: {
        defaultProps: {
          component: LinkBehavior,
          underline: 'none',
          color: 'inherit',
        } as LinkProps,
      },
      MuiChip: {
        defaultProps: {
          variant: 'filled',
          deleteIcon: <ClearOutlined />,
        },
        styleOverrides: {
          root: {
            borderRadius: 4,
            backgroundColor: theme.palette.background.default,
          },
        },
      },
      MuiButtonBase: {
        defaultProps: {
          disableRipple: true,
          LinkComponent: LinkBehavior,
        },
      },
      MuiButton: {
        styleOverrides: {
          contained: {
            top: theme.spacing(0.5),
            padding: theme.spacing(1),
            margin: theme.spacing(2, 0, 2, 0),
          },
        },
        defaultProps: {
          disableElevation: true,
          LinkComponent: LinkBehavior,
        },
      },
      MuiIconButton: {
        defaultProps: {
          disableRipple: true,
        },
        styleOverrides: {
          root: {
            '&:hover': {
              background: 'none',
              color: theme.palette.primary.main,
            },
          },
        },
      },
      MuiInputLabel: {
        defaultProps: {
          shrink: true,
        },
        styleOverrides: {
          root: {
            textTransform: 'uppercase',
            fontSize: '1rem',
          },
          asterisk: {
            color: 'darkred',
            fontWeight: 'bold',
            fontSize: 'x-large',
          },
        },
      },
      MuiTextField: {
        defaultProps: {
          variant: 'standard',
        },
      },
      MuiFormControl: {
        defaultProps: {
          variant: 'standard',
        },
      },
      MuiInput: {
        defaultProps: {
          disableUnderline: true,
        },
        styleOverrides: {
          root: {
            border: `1px solid ${grey['300']}`,
            borderRadius: theme.shape.borderRadius,
            top: theme.spacing(0.5),
            padding: theme.spacing(1),
            margin: theme.spacing(2, 0, 2, 0),
            outline: `1px solid transparent`,
            '&.Mui-focused': {
              border: `1px solid ${theme.palette.primary.main}`,
              outline: `1px solid ${theme.palette.primary.main}`,
            },
            '&.Mui-error': {
              border: `1px solid ${theme.palette.error.main}`,
              outline: `1px solid ${theme.palette.error.main}`,
            },
          },
        },
      },
      MuiFormHelperText: {
        styleOverrides: {
          root: {
            marginTop: theme.spacing(-1),
          },
        },
      },
      MuiInputAdornment: {
        styleOverrides: {
          positionEnd: {
            '&.MuiIconButton': {
              marginRight: 0,
            },
          },
        },
      },
      MuiAutocomplete: {
        styleOverrides: {
          inputRoot: {
            paddingBottom: theme.spacing(1),
          },
        },
      },
      MuiPaper: {
        defaultProps: {
          variant: 'elevation',
          elevation: 0,
        },
      },
      MuiSnackbar: {
        defaultProps: {
          anchorOrigin: {
            horizontal: 'center',
            vertical: 'bottom',
          },
        },
        styleOverrides: {
          root: {},
        },
      },
      MuiSnackbarContent: {
        defaultProps: {
          elevation: 0,
        },
      },
      MuiAlert: {},
      MuiAvatar: {
        defaultProps: {
          variant: 'square',
        },
        styleOverrides: {
          root: {
            borderRadius: 12,
          },
        },
      },
      MuiTable: {
        styleOverrides: {
          root: {
            marginTop: theme.spacing(-1),
            borderCollapse: 'separate',
            // if borderCollapse: "separate",
            borderSpacing: theme.spacing(0, 1),
            padding: theme.spacing(1),
            paddingTop: 0,
            paddingBottom: 0,
          },
        },
      },
      MuiTableHead: {
        styleOverrides: {
          root: {
            textTransform: 'uppercase',
          },
        },
      },
      MuiTableBody: {
        styleOverrides: {
          root: {},
        },
      },
      MuiTableRow: {
        styleOverrides: {
          root: {
            backgroundColor: theme.palette.background.paper,
            td: {
              // if borderCollapse: "collapse",
              // borderTop: `solid ${theme.palette.background.default} 10px`,
              // borderBottom: `solid ${theme.palette.background.default} 10px`,
              '&:first-of-type': {
                borderTopLeftRadius: `${theme.shape.borderRadius}px`,
                borderBottomLeftRadius: `${theme.shape.borderRadius}px`,
              },
              '&:last-child': {
                borderTopRightRadius: `${theme.shape.borderRadius}px`,
                borderBottomRightRadius: `${theme.shape.borderRadius}px`,
              },
            },
          },
        },
      },
      MuiTableCell: {
        styleOverrides: {
          root: {
            border: 'none',
          },
        },
      },
    },
  }
}
