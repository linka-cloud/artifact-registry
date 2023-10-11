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

import React, { PropsWithChildren, useContext, useEffect, useState } from 'react'
import { useAsync, usePersistedState } from '../hooks'
import { AsyncVoid, Void } from '../utils'
import { API as _API, api } from './api'
import { Credentials } from './schemas/login'

type API = Omit<_API, 'login' | 'logout' | 'credentials'>

interface APIContext extends API {
  login: (credentials: Credentials, repo: string) => Promise<[boolean, Error?]>;
  logout: () => Promise<void>;
  credentials?: Credentials;
  authenticated?: boolean;
  baseRepo?: string;
  setBaseRepo: (repo?: string) => void;
}


const apiContext = React.createContext<APIContext>({
  ...api,
  login: async () => [false],
  logout: AsyncVoid,
  credentials: undefined,
  setBaseRepo: Void,
})


export interface APIProviderProps extends PropsWithChildren<any> {
  user?: string;
  password?: string;
}

export const APIProvider = ({ children }: APIProviderProps) => {
  const [baseRepo, setBaseRepo] = usePersistedState<string | undefined>(undefined, 'baseRepo')
  const [authenticated, setAuthenticated, loaded] = usePersistedState<boolean|undefined>(false, 'authenticated')
  const [credentials, setCredentials] = useState<Credentials>()

  useAsync(async () => {
    if (!authenticated) {
      setCredentials(undefined)
      return
    }
    const [{user, password}, error] = await api.credentials()
    if (error) {
      console.error(error)
      setCredentials(undefined)
      return
    }
    setCredentials({user: user!!, password: password!!})
  }, [authenticated])

  useAsync(async () => {
    if (!loaded || authenticated) return
    // check if auth is required
    const [_, error] = await api.repositories(baseRepo)
    if (error) {
      setAuthenticated(false)
      return
    }
    setAuthenticated(true)
  }, [loaded, authenticated])

  const login = async ({ user, password }: Credentials, repo: string = '') => {
    const [success, error] = await api.login(user, password, repo)
    if (success) setAuthenticated(true)
    return [success, error] as [boolean, Error?]
  }
  const logout = async () => {
    await api.logout()
    setAuthenticated(undefined)
    setBaseRepo(undefined)
  }
  return <apiContext.Provider
    value={{
      ...api,
      credentials,
      login,
      logout,
      authenticated: loaded ? authenticated : undefined,
      baseRepo,
      setBaseRepo,
    }}>{children}</apiContext.Provider>
}

export const useAPI = () => {
  const api = useContext(apiContext)
  return api
}

