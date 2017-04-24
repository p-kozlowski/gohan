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

package goext

import "encoding/json"

// IEnvironment is the only scope of Gohan available for a golang extensions;
// other packages must not be imported nor used
type IEnvironment interface {
	// modules
	Core() ICore
	Logger() ILogger
	Schemas() ISchemas
	Sync() ISync
	Database() IDatabase

	// state
	Reset()
}

// IEnvironmentSupport indicates that an object holds a reference to a valid environment
type IEnvironmentSupport interface {
	Environment() IEnvironment
}

type NullString struct {
	String string
	Valid  bool
}

type NullBool struct {
	Bool  bool
	Valid bool
}

type NullInt struct {
	Int   int
	Valid bool
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.String)
	}
	return json.Marshal(false)
}

func (ns *NullString) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		var valid bool
		if err := json.Unmarshal(b, &valid); err != nil {
			return err
		}
		ns.Valid = valid
		return nil
	}
	ns.String = s
	ns.Valid = true
	return nil
}

func (nb NullBool) MarshalJSON() ([]byte, error) {
	if nb.Valid {
		return json.Marshal(nb.Bool)
	}
	return json.Marshal(false)
}

func (nb *NullBool) UnmarshalJSON(b []byte) error {
	var val bool
	if err := json.Unmarshal(b, &val); err != nil {
		var valid bool
		if err := json.Unmarshal(b, &valid); err != nil {
			return err
		}
		nb.Valid = valid
		return nil
	}
	nb.Bool = val
	nb.Valid = true
	return nil
}

func (ni NullInt) MarshalJSON() ([]byte, error) {
	if ni.Valid {
		return json.Marshal(ni.Int)
	}
	return json.Marshal(false)
}

func (ni *NullInt) UnmarshalJSON(b []byte) error {
	var i int
	if err := json.Unmarshal(b, &i); err != nil {
		var valid bool
		if err := json.Unmarshal(b, &valid); err != nil {
			return err
		}
		ni.Valid = valid
		return nil
	}
	ni.Int = i
	ni.Valid = true
	return nil
}
