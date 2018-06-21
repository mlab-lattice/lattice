package mock

import (
	"github.com/mlab-lattice/lattice/pkg/definition"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/util/git"
)

type MockSystemResolver struct {
}

func newMockSystemResolver() *MockSystemResolver {
	return &MockSystemResolver{}
}

func (*MockSystemResolver) ResolveDefinition(uri string, gitResolveOptions *git.Options) (tree.Node, error) {
	return getMockSystemDefinition()
}

func (*MockSystemResolver) ListDefinitionVersions(uri string, gitResolveOptions *git.Options) ([]string, error) {
	return []string{"1.0.0", "2.0.0"}, nil
}

func getMockSystemDefinition() (tree.Node, error) {
	jsonBytes := []byte(mockSystemJson)

	defInterface, err := definition.NewFromJSON(jsonBytes)

	if err != nil {
		return nil, err
	}

	return tree.NewNode(defInterface, nil)
}

const mockSystemJson = `
{
  "type": "system",
  "name": "mock-system",
  "subsystems": [
    {
      "type": "service",
      "name": "api",
      "description": "Backend API for Petflix app",
      "components": [
        {
          "name": "api",
          "ports": [
            {
              "name": "http",
              "port": 80,
              "protocol": "HTTP"
            }
          ],
          "build": {
            "git_repository": {
              "url": "https://github.com/mlab-lattice/example-petflix-service",
              "commit": "08c58b1e42542af92c7078c8d345f09d00e1eb17"
            },
            "language": "node:boron",
            "command": "npm install"
          },
          "exec": {
            "command": [
              "node",
              "index.js"
            ],
            "environment": {
              "MONGODB_URI": "mongodb://lattice:lattice@ds161873.mlab.com:61873/tim-demo",
              "PORT": "80"
            }
          }
        }
      ],
      "resources": {
        "num_instances": 1,
        "instance_type": "t2.small"
      }
    },
    {
      "type": "service",
      "name": "www",
      "description": "Webserver for public web app. Serves static files for client and proxies to API",
      "components": [
        {
          "name": "www",
          "ports": [
            {
              "name": "http",
              "port": 8080,
              "protocol": "HTTP",
              "external_access": {
                "public": true
              }
            }
          ],
          "build": {
            "git_repository": {
              "url": "https://github.com/mlab-lattice/example-petflix-www",
              "commit": "398d7c66ad4c5c012a7d7b4ba3008ed6d86075a3"
            },
            "language": "node:boron",
            "command": "npm install && npm run build"
          },
          "exec": {
            "command": [
              "node",
              "app.js"
            ],
            "environment": {
              "PETFLIX_API_URI": "http://api.petflix.local"
            }
          }
        }
      ],
      "resources": {
        "num_instances": 1,
        "instance_type": "t2.small"
      }
    }
  ]
}

`
