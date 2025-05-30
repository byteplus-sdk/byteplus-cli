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

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// ShowJson print data as json
// data should be map[string]interface{}
func ShowJson(data interface{}, color bool) {
	if color {
		colorfulJson(data, 0, false, true)
	} else {
		buf := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(buf)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "    ")
		encoder.Encode(data)

		fmt.Println(buf.String())
	}
}

func colorfulJson(data interface{}, indent int, indentValue, lastValue bool) {
	if data == nil {
		if !lastValue {
			printlnWithIndent(0, "\033[1;33mnull\033[0m,")
		} else {
			printlnWithIndent(0, "\033[1;33mnull\033[0m")
		}
		return
	}

	switch v := data.(type) {
	case map[string]interface{}:
		if !indentValue {
			printlnWithIndent(0, "{")
		} else {
			printlnWithIndent(indent, "{")
		}
		defer func() {
			printWithIndent(indent, "}")
			if !lastValue {
				fmt.Print(",\n")
			} else {
				fmt.Print("\n")
			}
		}()

		loop, mapLen := 1, len(v)
		for k1, v1 := range v {
			printfWithIndent(indent+1, "\033[1;35m%q\033[0m", k1)
			fmt.Print(": ")
			colorfulJson(v1, indent+1, false, loop == mapLen)
			loop++
		}
	case []interface{}:
		if !indentValue {
			printlnWithIndent(0, "[")
		} else {
			printlnWithIndent(indent, "[")
		}
		defer func() {
			printWithIndent(indent, "]")
			if !lastValue {
				fmt.Print(",\n")
			} else {
				fmt.Print("\n")
			}
		}()

		loop, arrLen := 1, len(v)
		for _, v1 := range v {
			colorfulJson(v1, indent+1, true, loop == arrLen)
			loop++
		}
	case string:
		if indentValue {
			printfWithIndent(indent, "\033[1;32m%q\033[0m", v)
		} else {
			printfWithIndent(0, "\033[1;32m%q\033[0m", v)
		}
		if !lastValue {
			fmt.Print(",\n")
		} else {
			fmt.Print("\n")
		}
	case json.Number:
		if indentValue {
			printfWithIndent(indent, "\033[1;94m%v\033[0m", v)
		} else {
			printfWithIndent(0, "\033[1;94m%v\033[0m", v)
		}
		if !lastValue {
			fmt.Print(",\n")
		} else {
			fmt.Print("\n")
		}
	case bool:
		if indentValue {
			printfWithIndent(indent, "\033[1;91m%v\033[0m", v)
		} else {
			printfWithIndent(0, "\033[1;91m%v\033[0m", v)
		}
		if !lastValue {
			fmt.Print(",\n")
		} else {
			fmt.Print("\n")
		}
	default:
		if indentValue {
			printfWithIndent(indent, "\033[1;32m%v\033[0m", v)
		} else {
			printfWithIndent(0, "\033[1;32m%v\033[0m", v)
		}
		if !lastValue {
			fmt.Print(",\n")
		} else {
			fmt.Print("\n")
		}
	}
}

func printWithIndent(indent int, a ...interface{}) {
	for i := 0; i < 4*indent; i++ {
		fmt.Print(" ")
	}
	fmt.Print(a...)
}

func printlnWithIndent(indent int, a ...interface{}) {
	for i := 0; i < 4*indent; i++ {
		fmt.Print(" ")
	}
	fmt.Println(a...)
}

func printfWithIndent(indent int, format string, a ...interface{}) {
	for i := 0; i < 4*indent; i++ {
		fmt.Print(" ")
	}
	fmt.Printf(format, a...)
}
