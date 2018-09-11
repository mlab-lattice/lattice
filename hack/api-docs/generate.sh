# clone the docs repo if it does not exist

DOCS_DIR=$LATTICE_ROOT/pkg/api/docs
DOCS_REPO_NAME="lattice-api-docs"
DOCS_REPO_URI="git@github.com:mlab-lattice/$DOCS_REPO_NAME.git"

function clone_or_pull_docs_repo() {
    cd $DOCS_DIR
    # check if docs repo exists
    if [ ! -d $DOCS_REPO_NAME ]
    then
        git clone $DOCS_REPO_URI
    else
        cd $DOCS_REPO_NAME
        git pull
    fi
}


# clone or pull the docs repo
clone_or_pull_docs_repo

# generate swagger from go using swaggo/swag
# we need to cd into the directory first so that swag tool works properly especially with struct field descriptions
cd $LATTICE_ROOT/pkg/api && swag init -g server/rest/v1/api.go

# generate slate docs from swagger using widdershins
widdershins $LATTICE_ROOT/pkg/api/docs/swagger/swagger.yaml -o $LATTICE_ROOT/pkg/api/docs/lattice-api-docs/source/index.html.md --language_tabs "shell:shell"

# build static docs from slate using middleman
cd $LATTICE_ROOT/pkg/api/docs/lattice-api-docs && bundle exec middleman build --clean --build-dir ../build