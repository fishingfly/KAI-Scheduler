// Copyright 2025 NVIDIA CORPORATION
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NVIDIA/KAI-scheduler/pkg/binder/common"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	nvidiaGpusDir   = "/proc/driver/nvidia/gpus/"
	mthreadsGPUsDir = "/proc/driver/musa/"
)

func GetGPUDevice(ctx context.Context, podName string, namespace string) (string, error) {
	logger := log.FromContext(ctx)
	logger.Info("Getting GPU device id for pod", "namespace", namespace, "name", podName)

	hasNvidiaGpu, hasMthreadsGpu := delectGpuVendor()
	if hasNvidiaGpu {
		deviceSubDirs, err := os.ReadDir(nvidiaGpusDir)
		if err != nil {
			return "", fmt.Errorf("failed to read GPU devices dir: %v", err)
		}
		for _, subDir := range deviceSubDirs {
			infoFilePath := filepath.Join(nvidiaGpusDir, subDir.Name(), "information")

			logger.Info("Getting GPU device info", "path", infoFilePath)

			data, err := os.ReadFile(infoFilePath)
			if err != nil {
				return "", fmt.Errorf("failed to read GPU device information: %v", err)
			}

			content := string(data)
			logger.Info("GPU device info", "content", content)

			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.Contains(line, "GPU UUID") {
					words := strings.Fields(line)
					return words[len(words)-1], nil
				}
			}
		}
	} else if hasMthreadsGpu {
		// Find mthreads GPU UUID
		uuid, err := GetMthreadsGPUDevice()
		if err != nil {
			return "", fmt.Errorf("failed to find GPU UUID")
		}

		return uuid, nil
	}
	return "", fmt.Errorf("failed to find GPU UUID")
}

func GetMthreadsGPUDevice() (string, error) {
	gpuUuid, exist := os.LookupEnv(common.MthreadsVisibleDevices)
	if !exist {
		return "", fmt.Errorf("environment variable %s not found", common.MthreadsVisibleDevices)
	}

	return gpuUuid, nil
}

func delectGpuVendor() (bool, bool) {
	hasNvidiaGpu, hasMthreadsGpu := false, false
	if _, err := os.Stat(nvidiaGpusDir); err == nil {
		hasNvidiaGpu = true
	}
	if _, err := os.Stat(mthreadsGPUsDir); err == nil {
		hasMthreadsGpu = true
	}
	return hasNvidiaGpu, hasMthreadsGpu
}
