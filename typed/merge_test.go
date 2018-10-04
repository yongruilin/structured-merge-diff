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

package typed

import (
	"fmt"
	"reflect"
	"testing"

	"sigs.k8s.io/structured-merge-diff/schema"
	"sigs.k8s.io/structured-merge-diff/value"

	"gopkg.in/yaml.v2"
)

type mergeTestCase struct {
	name         string
	rootTypeName string
	schema       string
	triplets     []mergeTriplet
}

type mergeTriplet struct {
	lhs string
	rhs string
	out string
}

var mergeCases = []mergeTestCase{{
	name:         "simple pair",
	rootTypeName: "stringPair",
	schema: `types:
- name: stringPair
  struct:
    fields:
    - name: key
      type:
        scalar: string
    - name: value
      type:
        untyped: {}
`,
	triplets: []mergeTriplet{{
		`{"key":"foo","value":{}}`,
		`{"key":"foo","value":1}`,
		`{"key":"foo","value":1}`,
	}, {
		`{"key":"foo","value":{}}`,
		`{"key":"foo","value":1}`,
		`{"key":"foo","value":1}`,
	}, {
		`{"key":"foo","value":1}`,
		`{"key":"foo","value":{}}`,
		`{"key":"foo","value":{}}`,
	}, {
		`{"key":"foo","value":null}`,
		`{"key":"foo","value":{}}`,
		`{"key":"foo","value":{}}`,
	}, {
		`{"key":"foo"}`,
		`{"value":true}`,
		`{"key":"foo","value":true}`,
	}},
}, {
	name:         "null/empty map",
	rootTypeName: "nestedMap",
	schema: `types:
- name: nestedMap
  struct:
    fields:
    - name: inner
      type:
        map:
          elementType:
            untyped: {}
`,
	triplets: []mergeTriplet{{
		`{}`,
		`{"inner":{}}`,
		`{"inner":{}}`,
	}, {
		`{}`,
		`{"inner":null}`,
		`{"inner":null}`,
	}, {
		`{"inner":null}`,
		`{"inner":{}}`,
		`{"inner":{}}`,
	}, {
		`{"inner":{}}`,
		`{"inner":null}`,
		`{"inner":null}`,
	}, {
		`{"inner":{}}`,
		`{"inner":{}}`,
		`{"inner":{}}`,
	}},
}, {
	name:         "null/empty struct",
	rootTypeName: "nestedStruct",
	schema: `types:
- name: nestedStruct
  struct:
    fields:
    - name: inner
      type:
        struct:
          fields:
          - name: value
            type:
              untyped: {}
`,
	triplets: []mergeTriplet{{
		`{}`,
		`{"inner":{}}`,
		`{"inner":{}}`,
	}, {
		`{}`,
		`{"inner":null}`,
		`{"inner":null}`,
	}, {
		`{"inner":null}`,
		`{"inner":{}}`,
		`{"inner":{}}`,
	}, {
		`{"inner":{}}`,
		`{"inner":null}`,
		`{"inner":null}`,
	}, {
		`{"inner":{}}`,
		`{"inner":{}}`,
		`{"inner":{}}`,
	}},
}, {
	name:         "null/empty list",
	rootTypeName: "nestedList",
	schema: `types:
- name: nestedList
  struct:
    fields:
    - name: inner
      type:
        list:
          elementType:
            untyped: {}
          elementRelationship: atomic
`,
	triplets: []mergeTriplet{{
		`{}`,
		`{"inner":[]}`,
		`{"inner":[]}`,
	}, {
		`{}`,
		`{"inner":null}`,
		`{"inner":null}`,
	}, {
		`{"inner":null}`,
		`{"inner":[]}`,
		`{"inner":[]}`,
	}, {
		`{"inner":[]}`,
		`{"inner":null}`,
		`{"inner":null}`,
	}, {
		`{"inner":[]}`,
		`{"inner":[]}`,
		`{"inner":[]}`,
	}},
}, {
	name:         "struct grab bag",
	rootTypeName: "myStruct",
	schema: `types:
- name: myStruct
  struct:
    fields:
    - name: numeric
      type:
        scalar: numeric
    - name: string
      type:
        scalar: string
    - name: bool
      type:
        scalar: boolean
    - name: setStr
      type:
        list:
          elementType:
            scalar: string
          elementRelationship: associative
    - name: setBool
      type:
        list:
          elementType:
            scalar: boolean
          elementRelationship: associative
    - name: setNumeric
      type:
        list:
          elementType:
            scalar: numeric
          elementRelationship: associative
`,
	triplets: []mergeTriplet{{
		`{"numeric":1}`,
		`{"numeric":3.14159}`,
		`{"numeric":3.14159}`,
	}, {
		`{"numeric":3.14159}`,
		`{"numeric":1}`,
		`{"numeric":1}`,
	}, {
		`{"string":"aoeu"}`,
		`{"bool":true}`,
		`{"string":"aoeu","bool":true}`,
	}, {
		`{"setStr":["a","b","c"]}`,
		`{"setStr":["a","b"]}`,
		`{"setStr":["a","b","c"]}`,
	}, {
		`{"setStr":["a","b"]}`,
		`{"setStr":["a","b","c"]}`,
		`{"setStr":["a","b","c"]}`,
	}, {
		`{"setStr":["a","b","c"]}`,
		`{"setStr":[]}`,
		`{"setStr":["a","b","c"]}`,
	}, {
		`{"setStr":[]}`,
		`{"setStr":["a","b","c"]}`,
		`{"setStr":["a","b","c"]}`,
	}, {
		`{"setBool":[true]}`,
		`{"setBool":[false]}`,
		`{"setBool":[true,false]}`,
	}, {
		`{"setNumeric":[1,2,3.14159]}`,
		`{"setNumeric":[1,2,3]}`,
		// KNOWN BUG: this order is wrong
		`{"setNumeric":[1,2,3.14159,3]}`,
	}},
}, {
	name:         "associative list",
	rootTypeName: "myRoot",
	schema: `types:
- name: myRoot
  struct:
    fields:
    - name: list
      type:
        namedType: myList
    - name: atomicList
      type:
        namedType: mySequence
- name: myList
  list:
    elementType:
      namedType: myElement
    elementRelationship: associative
    keys:
    - key
    - id
- name: mySequence
  list:
    elementType:
      scalar: string
    elementRelationship: atomic
- name: myElement
  struct:
    fields:
    - name: key
      type:
        scalar: string
    - name: id
      type:
        scalar: numeric
    - name: value
      type:
        namedType: myValue
    - name: bv
      type:
        scalar: boolean
    - name: nv
      type:
        scalar: numeric
- name: myValue
  map:
    elementType:
      scalar: string
`,
	triplets: []mergeTriplet{{
		`{"list":[{"key":"a","id":1,"value":{"a":"a"}}]}`,
		`{"list":[{"key":"a","id":1,"value":{"a":"a"}}]}`,
		`{"list":[{"key":"a","id":1,"value":{"a":"a"}}]}`,
	}, {
		`{"list":[{"key":"a","id":1,"value":{"a":"a"}}]}`,
		`{"list":[{"key":"a","id":2,"value":{"a":"a"}}]}`,
		`{"list":[{"key":"a","id":1,"value":{"a":"a"}},{"key":"a","id":2,"value":{"a":"a"}}]}`,
	}, {
		`{"list":[{"key":"a","id":1},{"key":"b","id":1}]}`,
		`{"list":[{"key":"a","id":1},{"key":"a","id":2}]}`,
		`{"list":[{"key":"a","id":1},{"key":"b","id":1},{"key":"a","id":2}]}`,
	}, {
		`{"atomicList":["a","a","a"]}`,
		`{"atomicList":null}`,
		`{"atomicList":null}`,
	}, {
		`{"atomicList":["a","b","c"]}`,
		`{"atomicList":[]}`,
		`{"atomicList":[]}`,
	}, {
		`{"atomicList":["a","a","a"]}`,
		`{"atomicList":["a","a"]}`,
		`{"atomicList":["a","a"]}`,
	}},
}}

func (tt mergeTestCase) test(t *testing.T) {
	var s schema.Schema
	err := yaml.Unmarshal([]byte(tt.schema), &s)
	if err != nil {
		t.Fatalf("unable to unmarshal schema")
	}

	for i, triplet := range tt.triplets {
		triplet := triplet
		t.Run(fmt.Sprintf("%v-valid-%v", tt.name, i), func(t *testing.T) {
			t.Parallel()

			lhs, err := value.FromYAML([]byte(triplet.lhs))
			if err != nil {
				t.Fatalf("unable to interpret lhs yaml: %v\n%v", err, triplet.lhs)
			}
			t.Logf("parsed lhs object:\v%v", lhs.HumanReadable())

			rhs, err := value.FromYAML([]byte(triplet.rhs))
			if err != nil {
				t.Fatalf("unable to interpret rhs yaml: %v\n%v", err, triplet.rhs)
			}
			t.Logf("parsed rhs object:\v%v", rhs.HumanReadable())

			expect, err := value.FromYAML([]byte(triplet.out))
			if err != nil {
				t.Fatalf("unable to interpret out yaml: %v\n%v", err, triplet.out)
			}
			t.Logf("parsed out object:\v%v", expect.HumanReadable())

			tvLHS := AsTypedUnvalidated(lhs, &s, tt.rootTypeName)
			tvRHS := AsTypedUnvalidated(rhs, &s, tt.rootTypeName)
			got, err := tvLHS.Merge(tvRHS)
			if err != nil {
				t.Errorf("got validation errors: %v", err)
			} else {
				t.Logf("got:\v%v", got.value.HumanReadable())
				gotUS := got.value.ToUnstructured(true)
				expectUS := expect.ToUnstructured(true)
				if !reflect.DeepEqual(gotUS, expectUS) {
					t.Errorf("Expected\n%v\nbut got\n%v\n",
						expect.HumanReadable(), got.value.HumanReadable(),
					)
				}
			}
		})
	}
}

func TestMerge(t *testing.T) {
	for _, tt := range mergeCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.test(t)
		})
	}
}
