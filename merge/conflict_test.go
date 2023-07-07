/*
Copyright 2018 The Kubernetes Authors.

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

package merge_test

import (
	"reflect"
	"testing"

	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
	"sigs.k8s.io/structured-merge-diff/v4/merge"
	"sigs.k8s.io/structured-merge-diff/v4/value"
)

var (
	// Short names for readable test cases.
	_NS  = fieldpath.NewSet
	_P   = fieldpath.MakePathOrDie
	_KBF = fieldpath.KeyByFields
	_V   = value.NewValueInterface
)

func TestNewFromSets(t *testing.T) {
	got := merge.ConflictsFromManagers(fieldpath.ManagedFields{
		"Bob": fieldpath.NewVersionedSet(
			_NS(
				_P("key"),
				_P("list", _KBF("key", "a", "id", 2), "id"),
			),
			"v1",
			false,
		),
		"Alice": fieldpath.NewVersionedSet(
			_NS(
				_P("value"),
				_P("list", _KBF("key", "a", "id", 2), "key"),
			),
			"v1",
			false,
		),
	})
	wanted := `conflicts with "Alice":
- .value
- .list[id=2,key="a"].key
conflicts with "Bob":
- .key
- .list[id=2,key="a"].id`
	if got.Error() != wanted {
		t.Errorf("Got %v, wanted %v", got.Error(), wanted)
	}
}

func TestToSet(t *testing.T) {
	conflicts := merge.ConflictsFromManagers(fieldpath.ManagedFields{
		"Bob": fieldpath.NewVersionedSet(
			_NS(
				_P("key"),
				_P("list", _KBF("key", "a", "id", 2), "id"),
			),
			"v1",
			false,
		),
		"Alice": fieldpath.NewVersionedSet(
			_NS(
				_P("value"),
				_P("list", _KBF("key", "a", "id", 2), "key"),
			),
			"v1",
			false,
		),
	})
	expected := fieldpath.NewSet(
		_P("key"),
		_P("value"),
		_P("list", _KBF("key", "a", "id", 2), "id"),
		_P("list", _KBF("key", "a", "id", 2), "key"),
	)
	actual := conflicts.ToSet()
	if !expected.Equals(actual) {
		t.Fatalf("expected\n%v\n, but got\n%v\n", expected, actual)
	}
}

func TestConflictsFromManagers(t *testing.T) {
	type args struct {
		sets fieldpath.ManagedFields
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test with common prefix",
			args: args{
				sets: fieldpath.ManagedFields{
					"Bob": fieldpath.NewVersionedSet(
						_NS(
							_P("spec", "template", "spec", "containers", _KBF("name", "probe"), "livenessProbe", "exec", "command"),
							_P("spec", "template", "spec", "containers", _KBF("name", "probe"), "livenessProbe", "periodSeconds"),
							_P("spec", "template", "spec", "containers", _KBF("name", "probe"), "readinessProbe", "exec", "command"),
							_P("spec", "template", "spec", "containers", _KBF("name", "probe"), "readinessProbe", "periodSeconds"),
						),
						"v1",
						false,
					),
				},
			},
			want: `conflicts with "Bob":
- .spec.template.spec.containers[name="probe"].livenessProbe.periodSeconds
- .spec.template.spec.containers[name="probe"].livenessProbe.exec.command
- .spec.template.spec.containers[name="probe"].readinessProbe.periodSeconds
- .spec.template.spec.containers[name="probe"].readinessProbe.exec.command`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := merge.ConflictsFromManagers(tt.args.sets); !reflect.DeepEqual(got.Error(), tt.want) {
				t.Errorf("ConflictsFromManagers() = %v, want %v", got, tt.want)
			}
		})
	}
}
