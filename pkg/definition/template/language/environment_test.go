package language

import (
	"testing"
)

func TestEnv(t *testing.T) {

	env := newEnvironment(NewEngine(), &Options{})

	if env.currentFrame() != nil {
		t.Fatalf("Current frame is not nill")
	}

	resource, _ := newURLResource("https://foo.bar/test.git/test.json",
		"https://foo.bar/test.git/",
		"test.json",
		make([]byte, 10))
	env.push(resource, map[string]interface{}{"x": 1}, map[string]interface{}{"y": 2})

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
