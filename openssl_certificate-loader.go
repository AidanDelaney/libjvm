/*
 * Copyright 2018-2020 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package libjvm

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

type OpenSSLCertificateLoader struct {
	DistributionType DistributionType
	JavaVersion      string
	LayerContributor libpak.HelperLayerContributor
	Logger           bard.Logger
}

func NewOpenSSLCertificateLoader(buildpack libcnb.Buildpack, distributionType DistributionType, javaVersion string,
	plan *libcnb.BuildpackPlan) OpenSSLCertificateLoader {

	layerContributor := libpak.NewHelperLayerContributor(filepath.Join(buildpack.Path, "bin", "openssl-certificate-loader"),
		"OpenSSL Certificate Loader", buildpack.Info, plan)
	layerContributor.LayerContributor.ExpectedMetadata = map[string]interface{}{
		"distribution-type": distributionType.String(),
		"info":              buildpack.Info,
		"java-version":      javaVersion,
	}

	return OpenSSLCertificateLoader{
		DistributionType: distributionType,
		JavaVersion:      javaVersion,
		LayerContributor: layerContributor,
	}
}

//go:generate statik -src . -include *.sh

func (o OpenSSLCertificateLoader) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	o.LayerContributor.Logger = o.Logger

	return o.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		o.Logger.Bodyf("Copying to %s", layer.Path)
		if err := sherpa.CopyFile(artifact, filepath.Join(layer.Path, "bin", "openssl-certificate-loader")); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to copy\n%w", err)
		}

		var source string
		if IsBeforeJava9(o.JavaVersion) && o.DistributionType == JDKType {
			source = filepath.Join("jre", "lib", "security", "cacerts")
		} else {
			source = filepath.Join("lib", "security", "cacerts")
		}

		s, err := sherpa.TemplateFile("/openssl-certificate-loader.sh", map[string]interface{}{"source": source})
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to load security-providers-configurer.sh\n%w", err)
		}

		layer.Profile.Add("openssl-certificate-loader.sh", s)

		layer.Launch = true
		return layer, nil
	})
}

func (o OpenSSLCertificateLoader) Name() string {
	return o.LayerContributor.LayerName()
}
