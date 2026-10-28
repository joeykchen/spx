/*
 * Copyright (c) 2021 The XGo Authors (xgo.dev). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ffi

/*
#include "gdextension_spx_interface.h"

void cgo_callfn_GDExtensionSpxGlobalFreeString(const GDExtensionSpxGlobalFreeString fn, GdString value) {
	fn(value);
}
void cgo_callfn_GDExtensionSpxGlobalRegisterCallbacks(const GDExtensionSpxGlobalRegisterCallbacks fn, GDExtensionSpxCallbackInfoPtr callback_ptr) {
	fn(callback_ptr);
}
*/
import "C"

var builtinAPI GDExtensionBuiltinInterface

type GDExtensionBuiltinInterface struct {
	SpxGlobalRegisterCallbacks GDExtensionSpxGlobalRegisterCallbacks
	SpxGlobalFreeString        GDExtensionSpxGlobalFreeString
}

func CallGlobalRegisterCallbacks(
	callback_ptr GDExtensionSpxCallbackInfoPtr,
) {
	arg0 := (C.GDExtensionSpxGlobalRegisterCallbacks)(builtinAPI.SpxGlobalRegisterCallbacks)
	arg1 := (C.GDExtensionSpxCallbackInfoPtr)(callback_ptr)

	C.cgo_callfn_GDExtensionSpxGlobalRegisterCallbacks(arg0, arg1)
}

func (x *GDExtensionBuiltinInterface) resolveAPIFunctions() {
	x.SpxGlobalRegisterCallbacks = (GDExtensionSpxGlobalRegisterCallbacks)(resolveCFunc("spx_global_register_callbacks"))
	x.SpxGlobalFreeString = (GDExtensionSpxGlobalFreeString)(resolveCFunc("spx_global_free_string"))
}

func freeReturnedString(value GdString) {
	C.cgo_callfn_GDExtensionSpxGlobalFreeString(
		(C.GDExtensionSpxGlobalFreeString)(builtinAPI.SpxGlobalFreeString),
		(C.GdString)(value),
	)
}
