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

import { Theme, useMediaQuery } from '@mui/material'
import { DependencyList, EffectCallback, useEffect, useState } from 'react'

export const useIsMobile = () => {
  return useMediaQuery<Theme>((theme) => theme.breakpoints.down('md'))
}

export const useEffectOnce = (effect: EffectCallback) => {
  useEffect(() => {
    return effect()
  }, [])
}

export const useAsync = (effect: () => Promise<void> | undefined, deps?: DependencyList) => {
  useEffect(() => {
    if (!effect) return
    // @ts-ignore
    effect().then()
  }, deps)
}

export const useAsyncOnce = (effect: () => Promise<void> | undefined) => {
  useAsync(effect, [])
}

export const usePersistedState = <T, >(defaultValue: T, key: string): [v: T, setV: (state: T) => void, loaded: boolean] => {
  const [loaded, setLoaded] = useState(false)
  const [state, _setState] = useState(defaultValue)
  useEffectOnce(() => {
    const s = localStorage.getItem(key)
    if (s) {
      setState(JSON.parse(s))
    }
    setLoaded(true)
  })
  const setState = (state: T) => {
    if (state === undefined) {
      localStorage.removeItem(key)
    } else {
      localStorage.setItem(key, JSON.stringify(state))
    }
    _setState(state)
  }
  return [state, setState, loaded]
}
