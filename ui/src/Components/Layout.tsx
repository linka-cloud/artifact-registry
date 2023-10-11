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

import {
  DarkModeOutlined,
  GitHub,
  LightModeOutlined,
  PowerSettingsNewOutlined,
  SettingsBrightnessOutlined,
} from '@mui/icons-material'
import {
  Box,
  Breadcrumbs,
  ButtonGroup,
  Container,
  IconButton,
  Link,
  LinkProps,
  Stack,
  StackProps,
  Tooltip,
  Typography,
} from '@mui/material'
import React, { useEffect, useRef } from 'react'
import { Link as RouterLink, useLocation } from 'react-router-dom'
import { useAPI } from '../api/useAPI'
import { useIsMobile } from '../hooks'
import LKARLogo from '../img/lkar-no-background.png'
import LinkaCloudLogo from '../img/linka-cloud-no-background-logo.png'
import { MainRoutesRegistry } from '../routes'
import { useUiMode } from '../theme/ColorModeProvider'
import { defaultPadding } from '../theme/theme'
import { ExternalLink } from './ExternalLink'

interface LinkRouterProps extends LinkProps {
  to: string;
  replace?: boolean;
}

const LinkRouter = (props: LinkRouterProps) => <Link {...props} component={RouterLink as any} />

export const Layout = ({ children }: React.PropsWithChildren<any>) => {
  const isMobile = useIsMobile()
  const location = useLocation()
  const locationRef = useRef(location.pathname)
  useEffect(() => {
    if (location.pathname === locationRef.current) {
      return
    }
    locationRef.current = location.pathname
    document.getElementById('main')?.scrollTo(0, 0)
  }, [location])
  return isMobile ? <MobileLayout>{children}</MobileLayout> : <TabletOrDesktopLayout>{children}</TabletOrDesktopLayout>
}

const Header = (props: StackProps) => {
  const { mode, setMode } = useUiMode()
  const location = useLocation()
  const pathnames = location.pathname.split('/').filter((x) => x)
  const routeLabel = (path: string) => Object.values(MainRoutesRegistry).find((r) => r.path === path)?.label
  return (
    <Stack component='header' paddingTop={defaultPadding} {...props}>
      <Stack direction='row' justifyContent='space-between'>
        <Box component='img' src={LKARLogo} width={100} sx={{padding: defaultPadding, paddingLeft: '0 !important'}}/>
        <ButtonGroup>
          <Tooltip title={`Switch to ${mode === 'light' ? 'Dark' : 'Light'} mode`}>
            <IconButton onClick={() => setMode(mode === 'light' ? 'dark' : 'light')}>
              {mode === 'light' ? <DarkModeOutlined /> : <LightModeOutlined />}
            </IconButton>
          </Tooltip>
          <Tooltip title={`Use system settings`}>
            <IconButton onClick={() => setMode(undefined)}>
              <SettingsBrightnessOutlined />
            </IconButton>
          </Tooltip>
          <Tooltip title={`logout`}>
            <IconButton component={Link} href='/logout'>
              <PowerSettingsNewOutlined />
            </IconButton>
          </Tooltip>
        </ButtonGroup>
      </Stack>
      <Breadcrumbs aria-label='breadcrumb'>
        <LinkRouter underline='hover' color='inherit' to='/'>
          {routeLabel('/')}
        </LinkRouter>
        {pathnames.map((value, index) => {
          const last = index === pathnames.length - 1
          const to = `/${pathnames.slice(0, index + 1).join('/')}`
          const label = routeLabel(to)
          return last ? (
            <Typography color='text.primary' key={to}>
              {label || decodeURIComponent(value)}
            </Typography>
          ) : (
            <LinkRouter underline='hover' color='inherit' to={to} key={to}>
              {label}
            </LinkRouter>
          )
        })}
      </Breadcrumbs>
    </Stack>
  )
}

const Footer = () => (
  <Stack
    component='footer'
    direction='row'
    alignItems='center'
    justifyContent='space-between'
    padding={theme => theme.spacing(1)}
    paddingTop={8}
  >
    <Box component='img' src={LinkaCloudLogo} flex={2} width={150} />
    <Typography variant='caption' color='text.secondary' flex={10} textAlign='center'>
      © 2023 Linka Cloud. All rights reserved.
    </Typography>
    <Stack sx={{ flex: 2 }}>
      <IconButton component={ExternalLink} href='https://github.com/linka-cloud/artifact-registry' sx={{alignSelf: "end", paddingRight: 0}}>
        <GitHub />
      </IconButton>
    </Stack>
  </Stack>
)

export const TabletOrDesktopLayout = ({ children }: React.PropsWithChildren<any>) => {
  const { authenticated } = useAPI()
  const { pathname } = useLocation()
  const show = authenticated && pathname !== '/logout'
  // const isXl = useMediaQuery<Theme>((theme) => theme.breakpoints.only("xl"));
  return show ? (
    <Stack id='main' component='main' minHeight='100%'>
      <Container sx={{
        display: 'flex',
        flexDirection: 'column',
        flex: 1,
        // marginLeft: isXl ? 0 : undefined,
        marginBottom: defaultPadding,
      }}>
        <Header padding={0} paddingLeft={1} paddingBottom={(theme) => theme.spacing(2)} />
        <Stack component='main' flex={1}>{children}</Stack>
        <Footer />
      </Container>
    </Stack>
  ) : (
    <UnauthenticatedLayout>{children}</UnauthenticatedLayout>
  )
}

export const UnauthenticatedLayout = ({ children }: React.PropsWithChildren<any>) => {
  const { mode, setMode } = useUiMode()
  return (
    <Stack minHeight='100%' padding={(theme) => theme.spacing(2)} paddingTop={defaultPadding}>
      <Container sx={{ display: 'flex', flexDirection: 'column', flex: 1 }}>
        <Stack direction='row' justifyContent='flex-end'>
          <ButtonGroup>
            <Tooltip title={`Switch to ${mode === 'light' ? 'Dark' : 'Light'} mode`}>
              <IconButton onClick={() => setMode(mode === 'light' ? 'dark' : 'light')}>
                {mode === 'light' ? <DarkModeOutlined /> : <LightModeOutlined />}
              </IconButton>
            </Tooltip>
            <Tooltip title={`Use system settings`}>
              <IconButton onClick={() => setMode(undefined)}>
                <SettingsBrightnessOutlined />
              </IconButton>
            </Tooltip>
          </ButtonGroup>
        </Stack>
        <Stack flex={1}>{children}</Stack>
      </Container>
    </Stack>
  )
}

export const MobileLayout = ({ children }: React.PropsWithChildren<any>) => {
  const { authenticated } = useAPI()
  const { pathname } = useLocation()
  const show = authenticated && pathname !== '/logout'

  return show ? (
    <>
      <Box component='main' id='main'>
        <Stack minHeight='calc(100% + 46px)' marginBottom='54px'>
          <Header padding={{ xs: 2, sm: 2 }} paddingBottom={0} />
          <Stack flex={1}>
            {children}
          </Stack>
          <Footer />
        </Stack>
      </Box>
    </>
  ) : (
    <UnauthenticatedLayout>{children}</UnauthenticatedLayout>
  )
}
