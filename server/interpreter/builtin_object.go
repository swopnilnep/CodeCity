/* Copyright 2017 Google Inc.
 * https://github.com/NeilFraser/CodeCity
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package interpreter

import (
	"fmt"

	"CodeCity/server/interpreter/data"
)

// This file contains code that creates the Object constructor and
// spec-specified properties on Object and Object.prototype, as well
// as providing native implementations for many of them.

func (intrp *Interpreter) initBuiltinObject() {
	// FIXME: Object should be constructor + conversion function.
	intrp.mkBuiltin("Object", data.NewObject(nil, intrp.protos.ObjectProto))

	intrp.mkBuiltin("Object.prototype", intrp.protos.ObjectProto)

	for _, ni := range builtinObjectNativeImpls {
		intrp.mkBuiltinFunc(ni.Tag)
	}
}

// Latin Letter Sinological Dot ('ꞏ', U+A78F) replaces '.' in names of
// builtin function implementations.

var builtinObjectNativeImpls = []NativeImpl{
	{"Object.getPrototypeOf", builtinObjectꞏgetPrototypeOf, 1},
	{"Object.getOwnPropertyDescriptor", builtinObjectꞏgetOwnPropertyDescriptor, 2},
	{"Object.getOwnPropertyNames", builtinObjectꞏgetOwnPropertyNames, 1},
	{"Object.create", builtinObjectꞏcreate, 2},
	{"Object.defineProperty", builtinObjectꞏdefineProperty, 3},
	{"Object.defineProperties", builtinObjectꞏdefineProperties, 2},
	// TODO(cpcallen): Finish Implementing §15.2.3 of ES5.1:
	// {"Object.seal", builtinObjectꞏseal, 1},
	// {"Object.freeze", builtinObjectꞏfreeze, 1},
	// {"Object.preventExtensions", builtinObjectꞏpreventExtensions, 1},
	// {"Object.isSealed", builtinObjectꞏisSealed, 1},
	// {"Object.isFrozen", builtinObjectꞏisFrozen, 1},
	// {"Object.isExtensible", builtinObjectꞏisExtensible, 1},
	// {"Object.keys", builtinObjectꞏkeys, 1},

	{"Object.prototype.toString", builtinObjectꞏprototypeꞏtoString, 0},
}

func builtinObjectꞏgetPrototypeOf(intrp *Interpreter, this data.Value, args []data.Value) (ret data.Value, throw bool) {
	obj, ok := args[0].(data.Object)
	if !ok {
		return intrp.typeError(fmt.Sprintf("Cannot get prototype of %s", args[0].ToString())), true
	}
	proto := obj.Proto()
	if proto == nil {
		return data.Null{}, false
	}
	return proto, false
}

func builtinObjectꞏgetOwnPropertyDescriptor(intrp *Interpreter, this data.Value, args []data.Value) (ret data.Value, throw bool) {
	obj, ok := args[0].(data.Object)
	if !ok {
		return intrp.typeError(fmt.Sprintf("Cannot get property descriptor from %s", args[0].ToString())), true
	}
	key := string(args[1].ToString())
	pd, ok := obj.GetOwnProperty(key)
	if !ok {
		return data.Undefined{}, false
	}
	// FIXME: set owner
	desc, nErr := data.FromPropertyDescriptor(pd, nil, intrp.protos.ObjectProto)
	if nErr != nil {
		return intrp.nativeError(nErr), true
	}
	return desc, false
}

func builtinObjectꞏgetOwnPropertyNames(intrp *Interpreter, this data.Value, args []data.Value) (ret data.Value, throw bool) {
	obj, ok := args[0].(data.Object)
	if !ok {
		return intrp.typeError(fmt.Sprintf("Cannot get propery names of %s", args[0].ToString())), true
	}
	keys := data.NewArray(nil, intrp.protos.ArrayProto)
	for i, k := range obj.OwnPropertyKeys() {
		nErr := keys.Set(string(data.Number(i).ToString()), data.String(k))
		if nErr != nil {
			return intrp.nativeError(nErr), true
		}
	}
	return keys, false
}

func builtinObjectꞏcreate(intrp *Interpreter, this data.Value, args []data.Value) (ret data.Value, throw bool) {
	if args[0] == (data.Null{}) {
		// FIXME: set owner
		return data.NewObject(nil, nil), false
	}
	proto, ok := args[0].(data.Object)
	if !ok {
		return intrp.typeError("Object prototype may only be an Object or null"), true
	}
	// FIXME: set owner
	obj := data.NewObject(nil, proto)
	if len(args) > 1 {
		builtinObjectꞏdefineProperties(intrp, this, []data.Value{obj, args[1]})
	}
	return obj, false

}

func builtinObjectꞏdefineProperty(intrp *Interpreter, this data.Value, args []data.Value) (ret data.Value, throw bool) {
	obj, ok := args[0].(data.Object)
	if !ok {
		return intrp.typeError(fmt.Sprintf("Cannot define property on %s", args[0].ToString())), true
	}
	key := string(args[1].ToString())
	desc, ok := args[2].(data.Object)
	if !ok {
		return intrp.typeError("Property descriptor must be an object"), true
	}
	pd, nErr := data.ToPropertyDescriptor(desc)
	if nErr != nil {
		return intrp.nativeError(nErr), true
	}
	nErr = obj.DefineOwnProperty(key, pd)
	if nErr != nil {
		return intrp.nativeError(nErr), true
	}
	return obj, false
}

func builtinObjectꞏdefineProperties(intrp *Interpreter, this data.Value, args []data.Value) (ret data.Value, throw bool) {
	obj, ok := args[0].(data.Object)
	if !ok {
		return intrp.typeError(fmt.Sprintf("Cannot define property on %s", args[0].ToString())), true
	}
	// FIXME: set owner:
	props, nErr := intrp.toObject(args[1], nil)
	if nErr != nil {
		return intrp.nativeError(nErr), true
	}
	type kpd struct {
		key string
		pd  data.Property
	}
	var kpds []kpd
	for _, key := range props.OwnPropertyKeys() {
		pdpd, ok := props.GetOwnProperty(key)
		if !ok || !pdpd.IsEnumerable() {
			continue
		}
		descObj, ok := pdpd.Value.(data.Object)
		if !ok {
			return intrp.typeError("Property descriptor must be an object"), true
		}
		pd, nErr := data.ToPropertyDescriptor(descObj)
		if nErr != nil {
			return intrp.nativeError(nErr), true
		}
		kpds = append(kpds, kpd{key, pd})
	}
	// Create props in second pass (in case of errors in first).
	for _, d := range kpds {
		nErr = obj.DefineOwnProperty(d.key, d.pd)
		if nErr != nil {
			return intrp.nativeError(nErr), true
		}
	}
	return obj, false
}

// TODO(cpcallen): Finish Implementing §15.2.3 of ES5.1:
// func builtinObjectꞏseal(intrp *Interpreter, this data.Value, args []data.Value) (ret data.Value, throw bool)
// func builtinObjectꞏfreeze(intrp *Interpreter, this data.Value, args []data.Value) (ret data.Value, throw bool)
// func builtinObjectꞏpreventExtensions(intrp *Interpreter, this data.Value, args []data.Value) (ret data.Value, throw bool)
// func builtinObjectꞏisSealed(intrp *Interpreter, this data.Value, args []data.Value) (ret data.Value, throw bool)
// func builtinObjectꞏisFrozen(intrp *Interpreter, this data.Value, args []data.Value) (ret data.Value, throw bool)
// func builtinObjectꞏisExtensible(intrp *Interpreter, this data.Value, args []data.Value) (ret data.Value, throw bool)
// func builtinObjectꞏkeys(intrp *Interpreter, this data.Value, args []data.Value) (ret data.Value, throw bool)

/****************************************************************/

func builtinObjectꞏprototypeꞏtoString(intrp *Interpreter, this data.Value, args []data.Value) (ret data.Value, throw bool) {
	// FIXME: don't panic
	if this == nil {
		panic("Object.property.toString called with this == nil??")
	}

	// FIXME: this is not quite to spec.  E.g.: applied to a new
	// String, it should return [object String] rather than the string
	// value.
	return this.ToString(), false
}

func init() {
	for _, ni := range builtinObjectNativeImpls {
		registerNativeImpl(ni)
	}
}
