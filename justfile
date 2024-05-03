default:
  just --list

# run tests for a module
test module test='all':
    dagger call -m {{module}}/tests {{env('ARGS', '')}} {{test}}

# tag and release a module
release module bump='minor':
    #!/usr/bin/env bash
    set -euo pipefail

    git checkout main > /dev/null 2>&1
    git diff-index --quiet HEAD || (echo "Git directory is dirty" && exit 1)

    version=v$(semver bump {{bump}} $(git tag | grep -oP "^{{module}}/\K.*" | sort | tail -1))

    echo "Tagging \"{{module}}\" module with version ${version}"
    read -n 1 -p "Proceed (y/N)? " answer
    echo

    case ${answer:0:1} in
        y|Y )
        ;;
        * )
            echo "Aborting"
            exit 1
        ;;
    esac

    tag={{module}}/$version

    git tag -m "{{module}}: ${version}" $tag
    git push origin $tag
