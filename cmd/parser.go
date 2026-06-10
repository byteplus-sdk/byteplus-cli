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
	"strings"
)

type Parser struct {
	currentIndex int
	args         []string
	currentFlag  *Flag
}

func NewParser(args []string) *Parser {
	return &Parser{
		args:         args,
		currentIndex: 0,
		currentFlag:  nil,
	}
}

func (p *Parser) ReadArgs(ctx *Context) ([]string, error) {
	if ctx == nil || ctx.fixedFlags == nil || ctx.dynamicFlags == nil {
		return nil, fmt.Errorf("invalid context for parsing arguments")
	}

	var r []string
	for {
		arg, _, more, err := p.readArg(ctx)
		if err != nil {
			return r, err
		}
		if arg != "" {
			r = append(r, arg)
		}
		if !more {
			return r, nil
		}
	}
}

func (p *Parser) readArg(ctx *Context) (arg string, flag *Flag, more bool, err error) {
	if len(p.args) <= p.currentIndex {
		if p.currentFlag != nil {
			err = p.currentFlagValueError(ctx)
			p.currentFlag = nil
		}
		more = false
		return
	}

	more = true
	rawArg := p.args[p.currentIndex]
	p.currentIndex++

	var value string
	flag, value, err = p.parseArg(rawArg, ctx)
	if err != nil {
		return
	}

	if p.currentFlag != nil && flag != nil {
		err = p.currentFlagValueError(ctx)
	}

	if flag == nil {
		if p.currentFlag != nil {
			if value == "" {
				err = p.currentFlagValueError(ctx)
			}
			p.currentFlag.SetValue(value)
			p.currentFlag = nil
		} else {
			arg = value
		}
	} else {
		p.currentFlag = flag
	}
	return
}

func (p *Parser) currentFlagValueError(ctx *Context) error {
	prefix := "--"
	if ctx != nil && ctx.fixedFlags != nil && ctx.fixedFlags.GetByName(p.currentFlag.Name) == p.currentFlag {
		prefix = "---"
	}
	return fmt.Errorf("%s%s must set value. ", prefix, p.currentFlag.Name)
}

func (p *Parser) parseArg(arg string, ctx *Context) (flag *Flag, value string, err error) {
	if strings.HasPrefix(arg, "---") {
		if len(arg) == 3 {
			err = fmt.Errorf("--- is not a valid flag")
		} else {
			// 三横线参数是 CLI 内部运行时覆盖参数，不参与 API 请求体构造。
			flag, err = ctx.fixedFlags.AddByName(arg[3:])
		}
	} else if strings.HasPrefix(arg, "--") {
		if len(arg) == 2 {
			err = fmt.Errorf("-- is not support command")
		} else {
			flag, err = ctx.dynamicFlags.AddByName(arg[2:])
		}
	} else {
		value = arg
	}
	return
}
