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

package util

// Copyright 2023 Byteplus.  All Rights Reserved.

import "fmt"

const (
	_BLACK   = "\033[30m"
	_RED     = "\033[31m"
	_GREEN   = "\033[32m"
	_YELLOW  = "\033[33m"
	_BLUE    = "\033[34m"
	_MAGENTA = "\033[35m"
	_CYAN    = "\033[36m"
	_WHITE   = "\033[37m"
	_DEFAULT = "\033[0m"
)

type colorPrinter struct {
	currentColor string
}

var cp colorPrinter

func setColor() {
	fmt.Print(cp.currentColor)
}

func resetColor() {
	fmt.Print(_DEFAULT)
}

func Black() *colorPrinter {
	cp.currentColor = _BLACK
	return &cp
}

func (cp *colorPrinter) Black() *colorPrinter {
	cp.currentColor = _BLACK
	return cp
}

func Red() *colorPrinter {
	cp.currentColor = _RED
	return &cp
}

func (cp *colorPrinter) Red() *colorPrinter {
	cp.currentColor = _RED
	return cp
}

func Green() *colorPrinter {
	cp.currentColor = _GREEN
	return &cp
}

func (cp *colorPrinter) Green() *colorPrinter {
	cp.currentColor = _GREEN
	return cp
}

func Yellow() *colorPrinter {
	cp.currentColor = _YELLOW
	return &cp
}

func (cp *colorPrinter) Yellow() *colorPrinter {
	cp.currentColor = _YELLOW
	return cp
}

func Blue() *colorPrinter {
	cp.currentColor = _BLUE
	return &cp
}

func (cp *colorPrinter) Blue() *colorPrinter {
	cp.currentColor = _BLUE
	return cp
}

func Magenta() *colorPrinter {
	cp.currentColor = _MAGENTA
	return &cp
}

func (cp *colorPrinter) Magenta() *colorPrinter {
	cp.currentColor = _MAGENTA
	return cp
}

func Cyan() *colorPrinter {
	cp.currentColor = _CYAN
	return &cp
}

func (cp *colorPrinter) Cyan() *colorPrinter {
	cp.currentColor = _CYAN
	return cp
}

func White() *colorPrinter {
	cp.currentColor = _WHITE
	return &cp
}

func (cp *colorPrinter) White() *colorPrinter {
	cp.currentColor = _WHITE
	return cp
}

func (cp *colorPrinter) Println(a ...interface{}) *colorPrinter {
	setColor()
	defer resetColor()
	fmt.Println(a...)
	return cp
}

func (cp *colorPrinter) Printf(format string, a ...interface{}) *colorPrinter {
	setColor()
	defer resetColor()
	fmt.Printf(format, a...)
	return cp
}

func (cp *colorPrinter) Print(a ...interface{}) *colorPrinter {
	fmt.Print(a...)
	return cp
}

func Println(a ...interface{}) *colorPrinter {
	fmt.Println(a...)
	return &cp
}

func Printf(format string, a ...interface{}) *colorPrinter {
	fmt.Printf(format, a...)
	return &cp
}

func Print(a ...interface{}) *colorPrinter {
	fmt.Print(a...)
	return &cp
}
