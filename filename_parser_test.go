/*
Copyright 2013 Google Inc.
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

package main

import (
	"testing"
)

func TestGetLocalIP(t *testing.T) {
	fn1 := FileName{name: "20170225T23:00:00Z_4.34.58.34_0.web100.gz"}
	if fn1.GetLocalIP() != "4.34.58.34" {
		t.Errorf("Wrong!\n")
		return
	}

	fn2 := FileName{name: "20170225T23:00:00Z_ALL0.web100.gz"}
	if fn2.GetLocalIP() != "" {
		t.Errorf("Wrong!\n")
		return
	}
}

func TestGetDate(t *testing.T) {
	fn1 := FileName{name: "20170225T23:00:00Z_4.34.58.34_0.web100.gz"}
	if fn1.GetDate() != "20170225" {
		t.Errorf("Wrong!\n")
		return
	}
}
