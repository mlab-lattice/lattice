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
	secretRef := SecretRef{
		Value: tree.PathSubcomponent("/foo/bar:buzz"),
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
				SecretRef: &secretRef,
			},
			d: []byte(`{"$secret_ref":"/foo/bar:buzz"}`),
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
	secretRef := SecretRef{
		Value: tree.PathSubcomponent("/foo/bar:buzz"),
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
				SecretRef: &secretRef,
			},
			d: []byte(`{"$secret_ref":"/foo/bar:buzz"}`),
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
