/*
Copyright 2018 The Fission Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"

	"github.com/fission/fission/pkg/canaryconfigmgr"
	config "github.com/fission/fission/pkg/featureconfig"
	"github.com/fission/fission/pkg/generated/clientset/versioned"
)

func ConfigCanaryFeature(ctx context.Context, logger *zap.Logger, fissionClient versioned.Interface, kubeClient kubernetes.Interface, featureConfig *config.FeatureConfig, featureStatus map[string]string) error {
	// start the appropriate controller
	if featureConfig.CanaryConfig.IsEnabled {
		canaryCfgMgr, err := canaryconfigmgr.MakeCanaryConfigMgr(ctx, logger, fissionClient, kubeClient, featureConfig.CanaryConfig.PrometheusSvc)
		if err != nil {
			featureStatus[config.CanaryFeature] = err.Error()
			return errors.Wrap(err, "failed to start canary config manager")
		}
		canaryCfgMgr.Run(ctx)
		logger.Info("started canary config manager")
	}

	return nil
}

// ConfigureFeatures gets the feature config and configures the features that are enabled
func ConfigureFeatures(ctx context.Context, logger *zap.Logger, unitTestMode bool, fissionClient versioned.Interface, kubeClient kubernetes.Interface) (map[string]string, error) {
	// set feature enabled to false if unitTestMode
	if unitTestMode {
		return nil, nil
	}

	// get the featureConfig from config map mounted onto the file system
	featureConfig, err := config.GetFeatureConfig()
	if err != nil {
		logger.Error("error getting feature config", zap.Error(err))
		return nil, err
	}

	featureStatus := make(map[string]string)

	// configure respective features
	// in the future when new optional features are added, we need to add corresponding feature handlers and invoke them here
	err = ConfigCanaryFeature(ctx, logger, fissionClient, kubeClient, featureConfig, featureStatus)
	return featureStatus, err
}
