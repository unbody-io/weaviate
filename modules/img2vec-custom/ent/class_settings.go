//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2024 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//

package ent

import (
	"errors"

	"github.com/weaviate/weaviate/entities/moduletools"
	basesettings "github.com/weaviate/weaviate/usecases/modulecomponents/settings"
)

type classSettings struct {
	basesettings.BaseClassSettings

	cfg moduletools.ClassConfig
}

func NewClassSettings(cfg moduletools.ClassConfig) *classSettings {
	return &classSettings{cfg: cfg}
}

func (ic *classSettings) Properties() ([]string, error) {
	if ic.cfg == nil {
		// we would receive a nil-config on cross-class requests, such as Explore{}
		return nil, errors.New("empty config")
	}

	imageFields, ok := ic.cfg.Class()["imageFields"]
	if !ok {
		return nil, errors.New("imageFields not present")
	}

	imageFieldsArray, ok := imageFields.([]interface{})
	if !ok {
		return nil, errors.New("imageFields must be an array")
	}

	fieldNames := make([]string, len(imageFieldsArray))
	for i, value := range imageFieldsArray {
		fieldNames[i] = value.(string)
	}
	return fieldNames, nil
}

func (ic *classSettings) ImageField(property string) bool {
	fieldNames, err := ic.Properties()
	if err != nil {
		return false
	}
	for i := range fieldNames {
		if fieldNames[i] == property {
			return true
		}
	}

	return false
}

func (ic *classSettings) Validate() error {
	return nil
}

func (ic *classSettings) GetEndpointURL() string {
	cls := ic.cfg.ClassByModuleName("img2vec-custom")

	for k, v := range cls {
		if k == "endpointUrl" || k == "endpointURL" {
			return v.(string)
		}
	}

	return ""
}

func (ic *classSettings) getProperty(name string) string {
	cls := ic.cfg.Class()
	if cls == nil {
		return ""
	}

	prop := cls[name]

	if prop != nil {
		return prop.(string)
	}

	return ""
}
