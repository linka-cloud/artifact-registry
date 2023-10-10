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

import { SvgIcon, SvgIconProps } from "@mui/material";

export const NotFoundIcon = (props: SvgIconProps) => (
  <SvgIcon viewBox="0 0 512 512" {...props} >
    <path d="M256,0C114.6,0,0,114.6,0,256s114.6,256,256,256s256-114.6,256-256S397.4,0,256,0z M64,256c0-106.1,86-192,192-192
	c42.1,0,81,13.7,112.6,36.7L100.7,368.6C77.7,337,64,298.1,64,256z M256,448c-42.1,0-81-13.7-112.6-36.7l267.9-267.9
	c23,31.7,36.7,70.5,36.7,112.6C448,362.1,362,448,256,448z"/>
  </SvgIcon>
);
