package v1

import (
	"encoding/json"
	"testing"
	//"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"reflect"
)

func TestValueOrSecretMarshalJSON(t *testing.T) {
	foo := "foo"
	path := tree.NodePathSubcomponent("/foo/bar:buzz")
	pathSecret := Secret{
		Path: &path,
	}
	localSecret := Secret{
		Local: &foo,
	}

	tests := []struct {
		v ValueOrSecret
		d []byte
	}{
		{
			v: ValueOrSecret{
				Value: &foo,
			},
			d: []byte(`"foo"`),
		},
		{
			v: ValueOrSecret{
				Secret: &pathSecret,
			},
			d: []byte(`{"$secret":"/foo/bar:buzz"}`),
		},
		{
			v: ValueOrSecret{
				Secret: &localSecret,
			},
			d: []byte(`{"$secret":"foo"}`),
		},
	}

	for _, test := range tests {
		d, err := json.Marshal(&test.v)
		if err != nil {
			t.Error(err)
			continue
		}

		if string(d) != string(test.d) {
			t.Errorf("expected %v but got %v", string(test.d), string(d))
		}
	}
}

func TestValueOrSecretUnmarshalJSON(t *testing.T) {
	foo := "foo"
	path := tree.NodePathSubcomponent("/foo/bar:buzz")
	pathSecret := Secret{
		Path: &path,
	}
	localSecret := Secret{
		Local: &foo,
	}

	tests := []struct {
		v ValueOrSecret
		d []byte
	}{
		{
			v: ValueOrSecret{
				Value: &foo,
			},
			d: []byte(`"foo"`),
		},
		{
			v: ValueOrSecret{
				Secret: &pathSecret,
			},
			d: []byte(`{"$secret":"/foo/bar:buzz"}`),
		},
		{
			v: ValueOrSecret{
				Secret: &localSecret,
			},
			d: []byte(`{"$secret":"foo"}`),
		},
	}

	for _, test := range tests {
		var v ValueOrSecret
		if err := json.Unmarshal(test.d, &v); err != nil {
			t.Error(err)
			continue
		}

		if !reflect.DeepEqual(v, test.v) {
			t.Errorf("expected %#v but got %#v", test.v, v)
		}
	}
}
