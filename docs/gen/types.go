// Copyright 2024 Linka Cloud  All rights reserved.
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

package main

import (
	"fmt"
	"strings"
)

type TemplateVariables struct {
	Domain     string
	DeployMode DeployMode
	RepoMode   RepoMode
}

func (d TemplateVariables) Registry(deployMode DeployMode, repoMode RepoMode, format, image string) string {
	var base string
	switch deployMode {
	case DeployModeSubdomain:
		base = fmt.Sprintf("%s.%s", format, d.Domain)
	case DeployModeSubPath:
		base = fmt.Sprintf("artifact-registry.%s", d.Domain)
	default:
	}
	if image == "" {
		return base
	}
	if repoMode == RepoModeMulti {
		return fmt.Sprintf("%s/%s", base, image)
	}
	return base
}

func (d TemplateVariables) RegistryURL(deployMode DeployMode, repoMode RepoMode, format, image string) string {
	base := d.Domain
	switch deployMode {
	case DeployModeSubdomain:
		base = fmt.Sprintf("%s.%s", format, base)
	case DeployModeSubPath:
		base = fmt.Sprintf("artifact-registry.%s/%s", base, format)
	default:
		return ""
	}
	if repoMode == RepoModeMulti {
		return fmt.Sprintf("%s/%s", base, image)
	}
	return base
}

func (d TemplateVariables) APIEndpoint(deployMode DeployMode, repoMode RepoMode, format, image, apiPath string) string {
	base := d.Domain
	if apiPath != "" {
		base = strings.Join([]string{d.Domain, apiPath}, "/")
	}
	switch deployMode {
	case DeployModeSubdomain:
		base = fmt.Sprintf("%s.%s", format, base)
	case DeployModeSubPath:
		base = fmt.Sprintf("artifact-registry.%s/%s", base, format)
	default:
		return ""
	}
	if repoMode == RepoModeMulti {
		return fmt.Sprintf("%s/%s", base, image)
	}
	return base
}

func (d TemplateVariables) DeployModes() []DeployMode {
	if d.DeployMode != 0 {
		return []DeployMode{d.DeployMode}
	}
	return []DeployMode{
		DeployModeSubPath,
		DeployModeSubdomain,
	}
}

func (d TemplateVariables) RepoModes() []RepoMode {
	if d.RepoMode != 0 {
		return []RepoMode{d.RepoMode}
	}
	return []RepoMode{
		RepoModeSingle,
		RepoModeMulti,
	}
}

func ParseDeployMode(s string) (DeployMode, error) {
	switch strings.ToLower(s) {
	case "subdomain":
		return DeployModeSubdomain, nil
	case "subpath":
		return DeployModeSubPath, nil
	case "":
		return DeployModeNone, nil
	default:
		return DeployModeNone, fmt.Errorf("unexpected deploy mode %q", s)
	}
}

type DeployMode uint

func (d DeployMode) String() string {
	switch d {
	case DeployModeSubdomain:
		return "Subdomain"
	case DeployModeSubPath:
		return "Subpath"
	default:
		return ""
	}
}

const (
	DeployModeNone DeployMode = iota
	DeployModeSubdomain
	DeployModeSubPath
)

func ParseRepoMode(s string) (RepoMode, error) {
	switch strings.ToLower(s) {
	case "single":
		return RepoModeSingle, nil
	case "multi":
		return RepoModeMulti, nil
	case "":
		return RepoModeNone, nil
	default:
		return RepoModeNone, fmt.Errorf("unexpected repo mode %q", s)
	}
}

type RepoMode uint

func (r RepoMode) String() string {
	switch r {
	case RepoModeSingle:
		return "Single"
	case RepoModeMulti:
		return "Multi"
	default:
		return ""
	}
}

const (
	RepoModeNone RepoMode = iota
	RepoModeSingle
	RepoModeMulti
)
