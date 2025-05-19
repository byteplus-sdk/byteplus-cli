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
	//跳出条件
	if len(p.args) <= p.currentIndex {
		more = false
		return
	}
	//设置下一跳
	more = true
	//获取当前位置的入参
	_arg := p.args[p.currentIndex]
	p.currentIndex++
	//计算是参数还是flag
	var (
		value string
	)
	flag, value, err = p.parseArg(_arg, ctx)
	if err != nil {
		return
	}

	//不允许两个连续的空--
	if p.currentFlag != nil && flag != nil {
		err = fmt.Errorf("--%s must set value. ", p.currentFlag.Name)
	}

	if flag == nil { //解析普通参数
		if p.currentFlag != nil {
			if value == "" {
				err = fmt.Errorf("--%s must set value. ", p.currentFlag.Name)
			}
			p.currentFlag.SetValue(value)
			p.currentFlag = nil
		} else {
			arg = value
		}
	} else { //解析flag
		p.currentFlag = flag
	}
	return
}

func (p *Parser) parseArg(arg string, ctx *Context) (flag *Flag, value string, err error) {
	if strings.HasPrefix(arg, "--") {
		if len(arg) == 2 {
			err = fmt.Errorf("-- is not support command")
		} else {
			//可变参数放入动态参数集合中
			flag, err = ctx.dynamicFlags.AddByName(arg[2:])
		}
	} else {
		value = arg
	}
	return
}
