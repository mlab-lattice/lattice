package sync

import (
	"testing"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"github.com/satori/go.uuid"
)

func TestLifecycleActions(t *testing.T) {
	type testPhase struct {
		description   string
		isTeardown    bool
		isRelease     bool
		deployID      v1.DeployID
		path          tree.Path
		teardownID    v1.TeardownID
		shouldSucceed bool
	}

	teardownE := v1.TeardownID("e")
	tests := []struct {
		description            string
		phases                 []testPhase
		expectedActiveDeploys  []v1.DeployID
		expectedActiveTeardown *v1.TeardownID
	}{
		{
			description:           "root",
			expectedActiveDeploys: []v1.DeployID{"c"},
			phases: []testPhase{
				{
					description:   "acquire root",
					deployID:      v1.DeployID("a"),
					path:          tree.RootPath(),
					shouldSucceed: true,
				},
				{
					description:   "reacquire root",
					deployID:      v1.DeployID("a"),
					path:          tree.RootPath(),
					shouldSucceed: true,
				},
				{
					description:   "different deploy fail to acquire root",
					deployID:      v1.DeployID("b"),
					path:          tree.RootPath(),
					shouldSucceed: false,
				},
				{
					description:   "teardown fail to acquire root",
					teardownID:    v1.TeardownID("c"),
					isTeardown:    true,
					shouldSucceed: false,
				},
				{
					description: "release root",
					deployID:    v1.DeployID("a"),
					isRelease:   true,
				},
				{
					description:   "different deploy acquire root",
					deployID:      v1.DeployID("c"),
					path:          tree.RootPath(),
					shouldSucceed: true,
				},
			},
		},
		{
			description:            "teardown",
			expectedActiveTeardown: &teardownE,
			phases: []testPhase{
				{
					description:   "acquire root",
					deployID:      v1.DeployID("a"),
					path:          tree.RootPath(),
					shouldSucceed: true,
				},
				{
					description: "release root",
					deployID:    v1.DeployID("a"),
					isRelease:   true,
				},
				{
					description:   "teardown acquire root",
					teardownID:    v1.TeardownID("c"),
					isTeardown:    true,
					shouldSucceed: true,
				},
				{
					description:   "original deploy fail to acquire root",
					deployID:      v1.DeployID("a"),
					path:          tree.RootPath(),
					shouldSucceed: false,
				},
				{
					description:   "different deploy fail to acquire root",
					deployID:      v1.DeployID("b"),
					path:          tree.RootPath(),
					shouldSucceed: false,
				},
				{
					description:   "different teardown fail to acquire root",
					deployID:      v1.DeployID("d"),
					path:          tree.RootPath(),
					shouldSucceed: false,
				},
				{
					description: "teardown release root",
					teardownID:  v1.TeardownID("c"),
					isRelease:   true,
					isTeardown:  true,
				},
				{
					description:   "different teardown acquire root",
					teardownID:    teardownE,
					isTeardown:    true,
					shouldSucceed: true,
				},
			},
		},
		{
			description:           "pathed deploys",
			expectedActiveDeploys: []v1.DeployID{"c", "d"},
			phases: []testPhase{
				{
					description:   "acquire root",
					deployID:      v1.DeployID("a"),
					path:          tree.RootPath(),
					shouldSucceed: true,
				},
				{
					description:   "different deploy fail to acquire subpath",
					deployID:      v1.DeployID("b"),
					path:          tree.RootPath().Child("b"),
					shouldSucceed: false,
				},
				{
					description:   "different deploy fail to acquire deeper subpath",
					deployID:      v1.DeployID("b"),
					path:          tree.RootPath().Child("b").Child("c"),
					shouldSucceed: false,
				},
				{
					description: "release root",
					deployID:    v1.DeployID("a"),
					isRelease:   true,
				},
				{
					description:   "acquire /a/b",
					deployID:      v1.DeployID("c"),
					path:          tree.RootPath().Child("a").Child("b"),
					shouldSucceed: true,
				},
				{
					description:   "acquire /a/c",
					deployID:      v1.DeployID("d"),
					path:          tree.RootPath().Child("a").Child("c"),
					shouldSucceed: true,
				},
				{
					description:   "fail to acquire /a/b/c",
					deployID:      v1.DeployID("e"),
					path:          tree.RootPath().Child("a").Child("b").Child("c"),
					shouldSucceed: false,
				},
				{
					description:   "fail to acquire /a",
					deployID:      v1.DeployID("f"),
					path:          tree.RootPath().Child("a"),
					shouldSucceed: false,
				},
			},
		},
	}

	namespace := v1.SystemID(uuid.NewV4().String())

	for _, test := range tests {
		passed := true
		actions := NewLifecycleActionManager()
		for _, phase := range test.phases {
			if phase.isRelease {
				if phase.isTeardown {
					actions.ReleaseTeardown(namespace, phase.teardownID)
				} else {
					actions.ReleaseDeploy(namespace, phase.deployID)
				}
				continue
			}

			var err error
			if phase.isTeardown {
				err = actions.AcquireTeardown(namespace, phase.teardownID)
			} else {
				err = actions.AcquireDeploy(namespace, phase.deployID, phase.path)
			}

			if phase.shouldSucceed && err != nil {
				t.Errorf("test %v failed. expected phase %v to succeed but got error: %v", test.description, phase.description, err)
				passed = false
				break
			} else if !phase.shouldSucceed && err == nil {
				t.Errorf("test %v failed. expected phase %v to fail but got no error", test.description, phase.description)
				passed = false
				break
			}
		}

		if passed {
			deploys, teardown := actions.InProgressActions(namespace)
			if len(deploys) != len(test.expectedActiveDeploys) {
				t.Errorf("test %v failed, expected %v active deploys but found %v", test.description, len(test.expectedActiveDeploys), len(deploys))
				continue
			}

			for _, expected := range test.expectedActiveDeploys {
				found := false
				for _, deploy := range deploys {
					if deploy == expected {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("test %v failed, expected %v to be an active deploy but it was not found", test.description, expected)
					continue
				}
			}

			if test.expectedActiveTeardown == nil && teardown != nil {
				t.Errorf("test %v failed, expected no active teardowns but found %v", test.description, *teardown)
			} else if test.expectedActiveTeardown != nil {
				if teardown == nil {
					t.Errorf("test %v failed, expected teardowns %v to be active but found none", test.description, *test.expectedActiveTeardown)
				} else if *teardown != *test.expectedActiveTeardown {
					t.Errorf("test %v failed, expected teardowns %v to be active but got %v", test.description, *test.expectedActiveTeardown, *teardown)
				}
			}
		}
	}
}
