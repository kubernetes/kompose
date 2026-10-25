/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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

package kubernetes

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	api "k8s.io/api/core/v1"
)

func TestInitConfigMapDataBinaryIsNotPreEncoded(t *testing.T) {
	raw := []byte{0x00, 0x01, 0xff, 0xfe, 'x'}
	cm := &api.ConfigMap{}
	initConfigMapData(cm, map[string]string{
		"blob.bin": string(raw),
		"note.txt": "hello\r\n",
	})

	if got := cm.BinaryData["blob.bin"]; string(got) != string(raw) {
		t.Fatalf("BinaryData = %v, want raw file bytes", got)
	}
	if _, ok := cm.Data["blob.bin"]; ok {
		t.Fatal("binary file was stored in Data")
	}
	if cm.Data["note.txt"] != "hello\n" {
		t.Fatalf("text Data = %q, want CRLF normalized", cm.Data["note.txt"])
	}

	encoded, err := json.Marshal(cm.BinaryData)
	if err != nil {
		t.Fatal(err)
	}
	var asStrings map[string]string
	if err := json.Unmarshal(encoded, &asStrings); err != nil {
		t.Fatal(err)
	}
	got, err := base64.StdEncoding.DecodeString(asStrings["blob.bin"])
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(raw) {
		t.Fatalf("marshaled binaryData decodes to %v, want %v", got, raw)
	}
}
