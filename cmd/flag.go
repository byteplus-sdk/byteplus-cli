/*
 * // Copyright (c) 2024 Bytedance Ltd. and/or its affiliates
 * //
 * // Licensed under the Apache License, Version 2.0 (the "License");
 * // you may not use this file except in compliance with the License.
 * // You may obtain a copy of the License at
 * //
 * //	http://www.apache.org/licenses/LICENSE-2.0
 * //
 * // Unless required by applicable law or agreed to in writing, software
 * // distributed under the License is distributed on an "AS IS" BASIS,
 * // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * // See the License for the specific language governing permissions and
 * // limitations under the License.
 */

package cmd

// Copyright 2023 Byteplus.  All Rights Reserved.

import (
	"fmt"
)

type Flag struct {
	Name  string
	value string
}

func (f *Flag) SetValue(value string) {
	f.value = value
}

func (f *Flag) GetValue() string {
	return f.value
}

type FlagSet struct {
	flags []*Flag
	index map[string]*Flag
}

func NewFlagSet() *FlagSet {
	return &FlagSet{
		flags: []*Flag{},
		index: make(map[string]*Flag),
	}
}

func (fs *FlagSet) GetFlags() []*Flag {
	return fs.flags
}

func (fs *FlagSet) AddFlag(f *Flag) {
	if f.Name != "" {
		key := "--" + f.Name
		if _, ok := fs.index[key]; ok {
			panic(fmt.Errorf("Flag is duplicated %s. ", key))
		}
		fs.index[key] = f
		fs.flags = append(fs.flags, f)
	}
}

func (fs *FlagSet) AddByName(name string) (*Flag, error) {
	f := &Flag{
		Name: name,
	}
	if _, ok := fs.index["--"+name]; ok {
		return nil, fmt.Errorf("flag duplicated --%s", name)
	}
	fs.AddFlag(f)
	return f, nil
}
