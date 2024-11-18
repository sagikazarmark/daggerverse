[private]
default:
  @just --list

# run tests for a module
test module test='all':
    dagger call -m {{module}}/tests {{env('ARGS', '')}} {{test}}

# tag and release a module
release module bump='minor':
    #!/usr/bin/env bash
    set -euo pipefail

    git checkout main > /dev/null 2>&1
    git diff-index --quiet HEAD || (echo "Git directory is dirty" && exit 1)

    version=v$(semver bump {{bump}} $(git tag --sort=v:refname | grep -oP "^{{module}}/\K.*" | tail -1 || echo "v0.0.0"))

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

# initialize a new module
[no-exit-message]
init module:
    @test ! -d {{module}} || (echo "Module \"{{module}}\" already exists" && exit 1)

    mkdir -p {{module}}
    cd {{module}} && dagger init --sdk go --name {{module}} --source .
    jq '.exclude = ["../.direnv", "../.devenv", "../go.work", "../go.work.sum", "tests"]' {{module}}/dagger.json | sponge {{module}}/dagger.json
    dagger develop -m {{module}}

    mkdir -p {{module}}/tests
    cd {{module}}/tests && dagger init --sdk go --name tests --source .
    jq '.exclude = ["../../.direnv", "../../.devenv", "../../go.work", "../../go.work.sum"]' {{module}}/tests/dagger.json | sponge {{module}}/tests/dagger.json
    go mod edit -module dagger/{{module}}/tests {{module}}/tests/go.mod
    cp -r .just/templates/tests/main.go {{module}}/tests/main.go
    cd {{module}}/tests && dagger install ..
    dagger develop -m {{module}}/tests

    @echo ""
    @echo "Module \"{{module}}\" initialized"
    @echo "Don't forget to add it to GitHub Actions when you are ready!"

# run `dagger develop` for all modules
develop:
    for dir in */; do dagger develop -m $dir; done
    for dir in */; do if [ -d "${dir%/}/tests" ]; then dagger develop -m "${dir%/}/tests"; fi done
    for dir in */; do if [ -d "${dir%/}/examples/go" ]; then dagger develop -m "${dir%/}/examples/go"; fi done
