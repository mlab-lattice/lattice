#!/usr/bin/env python

import sys
import traceback
import os
import shutil
import json

FORCE = False
########################################################################################################################
def load(a, b):
    # fake to get around the dependencies file
    pass
########################################################################################################################
def main(args):
    if not os.environ.get("GOPATH"):
        print "GOPATH is not set"
        exit(1)

    if "-f" in args:
        global FORCE
        FORCE = True

    print "Reading dependency tree"
    dependency_tree = get_dependencies()

    print "**** Dependency tree \n"
    print json.dumps(dependency_tree, indent=4)

    print "\n************************"

    copy_dependencies(dependency_tree)


""""/Users/abdul/github/mlab-lattice/src/github.com/mlab-lattice/lattice/bazel-lattice/external/com_github_tidwall_gjson/gjson.go"""

########################################################################################################################
def copy_dependencies(dependencies):
    if "importpath" in dependencies:
        copy_dependency_to_gopath(dependencies)
    else:
        for dep_path, dep in dependencies.items():
            # check if its a leaf
            copy_dependencies(dep)

########################################################################################################################

def copy_dependency_to_gopath(dependency):
    if "lattice" in dependency["name"]:
        print "Not copying lattice dependency %s" % dependency["name"]
        return

    src = get_dependency_path_in_bazel(dependency)
    dest = get_dependency_destination_gopath(dependency)

    try:
        if os.path.exists(dest):
            print "Already exists %s"\
                  % dest
            if FORCE:
                print "Replacing..."
                shutil.rmtree(dest)
            else:
                return
        print "Copying %s from '%s' to '%s'" % (dependency["name"], src, dest)
        shutil.copytree(src, dest)
    except Exception as ex:
        print "ERROR: %s" % ex

########################################################################################################################
def get_dependency_path_in_bazel(dependency):
    gopath = os.environ["GOPATH"]
    return os.path.join(gopath, "src/github.com/mlab-lattice/lattice/bazel-lattice/external", dependency["name"])

########################################################################################################################
def get_dependency_destination_gopath(dependency):
    gopath = os.environ["GOPATH"]
    return os.path.join(gopath, "src", dependency["importpath"])

########################################################################################################################
def get_dependencies():
    eval(compile(get_dependency_file_contents(), "<string>", 'exec'))
    return eval("GO_DEPENDENCIES")

########################################################################################################################
def get_dependency_file_contents():
    with open (get_dependency_file_path(), "r") as f:
        return f.read()

########################################################################################################################
def get_dependency_file_path():
    gopath = os.environ["GOPATH"]
    return os.path.join(gopath, "src/github.com/mlab-lattice/lattice/bazel/go/dependencies.bzl")

###############################################################################
########################                   ####################################
########################     BOOTSTRAP     ####################################
########################                   ####################################
###############################################################################


if __name__ == '__main__':
    try:

        main(sys.argv[1:])
    except (SystemExit, KeyboardInterrupt) , e:
        if hasattr(e, 'code') and e.code == 0:
            pass
        else:
            raise
    except:
        traceback.print_exc()