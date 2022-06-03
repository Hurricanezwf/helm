/*
 Copyright 2022 Hurricanezwf

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package helm

import (
	"context"
	"io"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"

	"github.com/pkg/errors"
)

type CMDInstall struct {
	// these fields are all required
	Namespace    string
	ReleaseName  string
	ChartPath    string
	Values       chartutil.Values
	Client       *action.Install
	Settings     *cli.EnvSettings
	Out          io.Writer
	DebugLogFunc DebugLogFunc
	WarnLogFunc  WarnLogFunc
}

func (c *CMDInstall) validateAndSetDefault() error {
	if c.Namespace == "" {
		return errors.New("namespace cannot be empty")
	}
	if c.ReleaseName == "" {
		return errors.New("release name cannot be empty")
	}
	if c.ChartPath == "" {
		return errors.New("chart path cannot be empty")
	}
	//if c.Values == nil {
	//	// it's ok
	//}
	if c.Client == nil {
		return errors.New("client cannot be nil")
	}
	if c.Settings == nil {
		c.Settings = cli.New() // use default
	}
	if c.Out == nil {
		c.Out = os.Stdout
	}
	if c.DebugLogFunc == nil {
		c.DebugLogFunc = DefaultDebugLogFunc()
	}
	if c.WarnLogFunc == nil {
		c.WarnLogFunc = DefaultWarnLogFunc()
	}
	return nil
}

func (c *CMDInstall) RunInstall(ctx context.Context) (*release.Release, error) {
	if err := c.validateAndSetDefault(); err != nil {
		return nil, err
	}

	// 转成 local 变量少改逻辑
	var (
		namespace   = c.Namespace
		releaseName = c.ReleaseName
		chartPath   = c.ChartPath
		values      = c.Values
		client      = c.Client
		settings    = c.Settings
		out         = c.Out
		debug       = c.DebugLogFunc
		warning     = c.WarnLogFunc
	)

	c.DebugLogFunc("Original chart version: %q", client.Version)
	if client.Version == "" && client.Devel {
		debug("setting version to >0.0.0-0")
		client.Version = ">0.0.0-0"
	}

	client.Namespace = namespace
	client.ReleaseName = releaseName

	cp, err := client.ChartPathOptions.LocateChart(chartPath, settings)
	if err != nil {
		return nil, err
	}

	debug("CHART PATH: %s\n", cp)

	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	if err := checkIfInstallable(chartRequested); err != nil {
		return nil, err
	}

	if chartRequested.Metadata.Deprecated {
		warning("This chart is deprecated")
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		p := getter.All(settings)

		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			err = errors.Wrap(err, "An error occurred while checking for chart dependencies. You may need to run `helm dependency build` to fetch missing dependencies")
			if client.DependencyUpdate {
				man := &downloader.Manager{
					Out:              out,
					ChartPath:        cp,
					Keyring:          client.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
					Debug:            settings.Debug,
				}
				if err := man.Update(); err != nil {
					return nil, err
				}
				// Reload the chart with the updated Chart.lock file.
				if chartRequested, err = loader.Load(cp); err != nil {
					return nil, errors.Wrap(err, "failed reloading chart after repo update")
				}
			} else {
				return nil, err
			}
		}
	}

	return client.RunWithContext(ctx, chartRequested, values)
}

// checkIfInstallable validates if a chart can be installed
//
// Application chart type is only installable
func checkIfInstallable(ch *chart.Chart) error {
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	return errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}
