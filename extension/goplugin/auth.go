// Copyright (C) 2017 NTT Innovation Institute, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package goplugin

import (
	"github.com/cloudwan/gohan/extension/goext"
	"github.com/cloudwan/gohan/schema"
)

// Auth is an implementation of IAuth
type Auth struct{}

func (a *Auth) HasRole(context goext.Context, principal string) bool {
	roleRaw, ok := context["role"]
	if !ok {
		log.Warning("HasRole: missing 'role' field in context")
		return false
	}

	role, ok := roleRaw.(*schema.Role)
	if !ok {
		log.Warning("HasRole: invalid type of 'role' field in context")
		return false
	}

	log.Warning("role name: %s, principal: %s, match %s", role.Name, principal, role.Match(principal))

	return role.Match(principal)
}

func (a *Auth) GetTenantName(context goext.Context) string {
	authRaw, ok := context["auth"]
	if !ok {
		log.Warning("GetTenantName: missing 'auth' field in context")
		return ""
	}

	auth, ok := authRaw.(schema.Authorization)
	if !ok {
		log.Warning("GetTenantName: invalid type of 'auth' field in context")
		return ""
	}

	log.Warning("tenant name: %s", auth.TenantName())
	return auth.TenantName()
}

// IsAdmin returns true if user had admin role
func (a *Auth) IsAdmin(context goext.Context) bool {
	return a.HasRole(context, "admin")
}
