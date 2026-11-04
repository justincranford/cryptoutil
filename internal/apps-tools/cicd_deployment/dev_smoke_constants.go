// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"time"
)

const (
	developmentServerField      = cryptoutilSharedMagic.CMD_SERVER
	developmentDefaultServer    = "https://127.0.0.1:6443"
	developmentKubeCurrent      = "current-context"
	developmentKubeClusters     = "clusters"
	developmentKubeContexts     = "contexts"
	developmentKubeContext      = "context"
	developmentKubeUser         = "user"
	developmentKubeUsers        = "users"
	developmentKubeCA           = "certificate-authority-data"
	developmentKubeCertificate  = "client-certificate-data"
	developmentKubeKey          = "client-key-data"
	developmentCommandField     = "command"
	developmentLoopback         = cryptoutilSharedMagic.IPv4Loopback
	developmentKubeconfigFlag   = "--kubeconfig"
	developmentContextFlag      = "--context"
	developmentVolume           = cryptoutilSharedMagic.DockerComposeVolumeMountTypeVolume
	developmentMaterialLifetime = 24 * time.Hour
	developmentCleanupTimeout   = 3 * time.Minute
	developmentRequestTimeout   = 15 * time.Second
	developmentCPUReserve       = 2000
	developmentMemoryReserve    = 2147483648
	developmentMaximumBody      = 1 << 20
	developmentFormatFlag       = "--format"
	developmentFileFlag         = "-f"
	developmentExec             = "exec"
	developmentDefaultContext   = "default"
	developmentNamespaceFlag    = "-n"
	developmentStdin            = "-"
	developmentContainer        = "container"
	developmentRemove           = "rm"
	developmentNetwork          = "network"
	developmentK3s              = "k3s"
	developmentNamespaceKind    = "Namespace"
	developmentMetadata         = "metadata"
	developmentName             = cryptoutilSharedMagic.ClaimName
	developmentLabels           = "labels"
	developmentImage            = "image"
	developmentAPIVersion       = "apiVersion"
	developmentKind             = "kind"
	developmentCreate           = "create"
	developmentInspect          = "inspect"
)
