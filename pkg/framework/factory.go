/*
Copyright 2021 The Kubernetes Authors.

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

package framework

import (
	"fmt"
	"sync"
)

// Factory is a default global factory instance.
var factory = newCnierFactory()

func newCnierFactory() *CniFactory {
	return &CniFactory{
		createFuncs: make(map[string]createCnierFunc),
	}
}

// CNIFactory is a factory that creates CNI instances.
type CniFactory struct {
	lock        sync.RWMutex
	createFuncs map[string]createCnierFunc
}

// Register registers create measurement function in measurement factory.
func Register(methodName string, createFunc createCnierFunc) error {
	return factory.register(methodName, createFunc)
}

func (cf *CniFactory) register(methodName string, createFunc createCnierFunc) error {
	cf.lock.Lock()
	defer cf.lock.Unlock()
	_, exists := cf.createFuncs[methodName]
	if exists {
		return fmt.Errorf("cni with method %v already exists", methodName)
	}
	cf.createFuncs[methodName] = createFunc
	return nil
}

func (cf *CniFactory) createCniInstance(cniName string) (Cnier, error) {
	cf.lock.RLock()
	defer cf.lock.RUnlock()
	createFunc, exists := cf.createFuncs[cniName]
	if !exists {
		return nil, fmt.Errorf("unknown cni %s", cniName)
	}
	return createFunc(), nil
}