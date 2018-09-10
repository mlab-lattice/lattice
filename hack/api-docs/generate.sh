# we need to cd into the directory first so that swag tool works properly especially with struct field descriptions
cd $LATTICE_ROOT/pkg/api && swag init -g server/rest/v1/api.go
widdershins $LATTICE_ROOT/pkg/api/docs/swagger/swagger.yaml -o $LATTICE_ROOT/pkg/api/docs/lattice-api-docs/source/index.html.md --language_tabs "shell:shell"
cd $LATTICE_ROOT/pkg/api/docs/lattice-api-docs && bundle exec middleman build --clean