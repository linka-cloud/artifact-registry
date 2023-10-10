// Copyright 2021 Linka Cloud  All rights reserved.
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


import React from 'react'

export const Edit = 'edit'
export const New = 'new'

export interface RouteDefinition {
  priority?: number;
  path: string;
  label?: string;
  component?: React.ReactNode;
  show?: boolean;
  icon?: React.ReactElement;
  hasBottomNavigation?: boolean,
  bottomEnd?: boolean
  subRoutes?: RoutesRegistry;
  public?: boolean
  navigate: (args?: any) => string
}

export interface RoutesRegistry {
  [key: string]: RouteDefinition;
}

export const MainRoutesRegistry: RoutesRegistry = {}
