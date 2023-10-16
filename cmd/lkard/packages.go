package main

import (
	_ "go.linka.cloud/artifact-registry/pkg/packages/apk"
	_ "go.linka.cloud/artifact-registry/pkg/packages/deb"
	_ "go.linka.cloud/artifact-registry/pkg/packages/helm"
	_ "go.linka.cloud/artifact-registry/pkg/packages/rpm"
)
