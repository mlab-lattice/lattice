package language

import (
	"testing"
)

func TestEnv(t *testing.T) {

	env := newEnvironment(NewEngine(), &Options{})

	if env.currentFrame() != nil {
		t.Fatalf("Current frame is not nill")
	}

	env.push("https://foo.bar/test.git/", map[string]interface{}{"x": 1}, map[string]interface{}{"y": 2})

	if env.currentFrame() == nil {
		t.Fatalf("Current frame is nill")
	}

	if env.currentFrame().variables == nil {
		t.Fatalf("Current frame variables is nill")
	}

	env.pop()

	if env.currentFrame() != nil {
		t.Fatalf("Current frame is not nill")
	}

}
