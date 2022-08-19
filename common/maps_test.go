package common_test

import (
	"reflect"
	"testing"

	"github.com/rd-benson/pigeon-service/common"
)

func TestMapDiffSlice(t *testing.T) {
	a := map[string][]string{
		"foo": {"foo1", "foo2"},
		"bar": {"bar1", "bar2"},
		"baz": {"baz1", "baz2"},
	}

	cases := []struct {
		description string
		b           map[string][]string
		wantAdd     map[string][]string
		wantRemove  map[string][]string
	}{
		{
			description: "identical maps",
			b:           a,
			wantAdd:     make(map[string][]string),
			wantRemove:  make(map[string][]string),
		},
		{
			description: "add multiple keys",
			b: map[string][]string{
				"foo":    {"foo1", "foo2"},
				"bar":    {"bar1", "bar2"},
				"baz":    {"baz1", "baz2"},
				"foobar": {"foobar1", "foobar2"},
				"foobaz": {"foobaz1", "foobaz2"},
			},
			wantAdd: map[string][]string{
				"foobar": {"foobar1", "foobar2"},
				"foobaz": {"foobaz1", "foobaz2"},
			},
			wantRemove: make(map[string][]string),
		},
		{
			description: "remove multiple keys",
			b: map[string][]string{
				"foo": {"foo1", "foo2"},
			},
			wantAdd: make(map[string][]string),
			wantRemove: map[string][]string{
				"bar": {"bar1", "bar2"},
				"baz": {"baz1", "baz2"},
			},
		},
		{
			description: "add values for multiple keys",
			b: map[string][]string{
				"foo": {"foo1", "foo2", "foo3"},
				"bar": {"bar1", "bar2", "bar3"},
				"baz": {"baz1", "baz2", "baz3"},
			},
			wantAdd: map[string][]string{
				"foo": {"foo3"},
				"bar": {"bar3"},
				"baz": {"baz3"},
			},
			wantRemove: make(map[string][]string),
		},
		{
			description: "add values for multiple keys",
			b: map[string][]string{
				"foo": {"foo1", "foo2", "foo3"},
				"bar": {"bar1", "bar2", "bar3"},
				"baz": {"baz1", "baz2", "baz3"},
			},
			wantAdd: map[string][]string{
				"foo": {"foo3"},
				"bar": {"bar3"},
				"baz": {"baz3"},
			},
			wantRemove: make(map[string][]string),
		},
		{
			description: "remove values for multiple keys",
			b: map[string][]string{
				"foo": {"foo1"},
				"bar": {"bar1"},
				"baz": {"baz1"},
			},
			wantAdd: make(map[string][]string),
			wantRemove: map[string][]string{
				"foo": {"foo2"},
				"bar": {"bar2"},
				"baz": {"baz2"},
			},
		},
	}

	for _, test := range cases {
		gotAdd, gotRemove := common.MapDiffSlice(a, test.b)
		assertEqualMapSlice(t, test.description+":add", gotAdd, test.wantAdd)
		assertEqualMapSlice(t, test.description+":remove", gotRemove, test.wantRemove)
	}

}

func assertEqualMapSlice[K comparable, V comparable](t *testing.T, description string, A map[K][]V, B map[K][]V) {
	t.Helper()
	if !reflect.DeepEqual(A, B) {
		t.Errorf("%v: got %v, wanted %v", description, A, B)
	}
}
