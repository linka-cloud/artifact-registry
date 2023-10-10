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

import { object, SchemaOf, string } from 'yup'

export interface Credentials {
  user: string
  password: string
}

export const credentialsSchema: SchemaOf<Credentials> = object({
  user: string().min(1).required('Required'),
  password: string().min(1).required('Required'),
})
