# clone the docs repo if it does not exist

DOCS_TARGET_DIR=$LATTICE_ROOT/api-docs
DOCS_REPO_NAME="lattice-api-docs"
DOCS_REPO_URI="git@github.com:mlab-lattice/$DOCS_REPO_NAME.git"
DOCS_REPO_DIR=$DOCS_TARGET_DIR/$DOCS_REPO_NAME

# create the docs target dir if it doesn't exist
if [ ! -d $DOCS_TARGET_DIR ]
then
    mkdir -p $DOCS_TARGET_DIR
fi


function clone_or_pull_docs_repo() {
    cd $DOCS_TARGET_DIR
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
cd $LATTICE_ROOT/pkg/api/server/rest && swagger generate spec -m -o $DOCS_TARGET_DIR/swagger.yaml

# generate slate docs from swagger using widdershins
widdershins $DOCS_TARGET_DIR/swagger.yaml -o $DOCS_REPO_DIR/source/index.html.md --language_tabs "shell:shell"

# build static docs from slate using middleman
cd $DOCS_REPO_DIR && bundle exec middleman build --clean --build-dir ../build