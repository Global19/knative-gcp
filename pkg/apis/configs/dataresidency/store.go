/*
Copyright 2020 Google LLC.

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

package dataresidency

import (
	"context"

	"knative.dev/pkg/configmap"
)

type dataresidencyCfgKey struct{}

// Config holds the collection of configurations that we attach to contexts.
// +k8s:deepcopy-gen=false
type Config struct {
	DataResidencyDefaults *Defaults
}

// FromContext extracts a Config from the provided context.
func FromContext(ctx context.Context) *Config {
	x, ok := ctx.Value(dataresidencyCfgKey{}).(*Config)
	if ok {
		return x
	}
	return nil
}

// FromContextOrDefaults is like FromContext, but when no Config is attached it
// returns a Config populated with the defaults for each of the Config fields.
func FromContextOrDefaults(ctx context.Context) *Config {
	if cfg := FromContext(ctx); cfg != nil {
		return cfg
	}
	defaults, _ := NewDefaultsConfigFromMap(map[string]string{})
	return &Config{
		DataResidencyDefaults: defaults,
	}
}

// ToContext attaches the provided Config to the provided context, returning the
// new context with the Config attached.
func ToContext(ctx context.Context, c *Config) context.Context {
	return context.WithValue(ctx, dataresidencyCfgKey{}, c)
}

// Store is a typed wrapper around configmap.Untyped store to handle our ConfigMaps.
// +k8s:deepcopy-gen=false
type Store struct {
	*configmap.UntypedStore
}

// NewStore creates a new store of Configs and optionally calls functions when ConfigMaps are updated.
func NewStore(logger configmap.Logger, onAfterStore ...func(name string, value interface{})) *Store {
	store := &Store{
		UntypedStore: configmap.NewUntypedStore(
			"data-residency-defaults",
			logger,
			configmap.Constructors{
				ConfigMapName(): NewDefaultsConfigFromConfigMap,
			},
			onAfterStore...,
		),
	}

	return store
}

// ToContext attaches the current Config state to the provided context.
func (s *Store) ToContext(ctx context.Context) context.Context {
	return ToContext(ctx, s.Load())
}

// Load creates a Config from the current config state of the Store.
func (s *Store) Load() *Config {
	return &Config{
		DataResidencyDefaults: s.UntypedLoad(ConfigMapName()).(*Defaults).DeepCopy(),
	}
}