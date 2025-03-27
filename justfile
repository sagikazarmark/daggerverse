[private]
default:
  @just --list

# initialize a new module
[no-exit-message]
init module:
    @test ! -d {{module}} || (echo "Module \"{{module}}\" already exists" && exit 1)

    mkdir -p {{module}}
    cd {{module}} && dagger init --sdk go --name {{module}} --source .
    jq '.include = ["!../.direnv", "!../.devenv", "!../go.work", "!../go.work.sum"]' {{module}}/dagger.json | sponge {{module}}/dagger.json
    dagger develop -m {{module}}

    mkdir -p {{module}}/tests
    cd {{module}}/tests && dagger init --sdk go --name tests --source .
    jq '.include = ["!../../.direnv", "!../../.devenv", "!../../go.work", "!../../go.work.sum"]' {{module}}/tests/dagger.json | sponge {{module}}/tests/dagger.json
    go mod edit -module dagger/{{module}}/tests {{module}}/tests/go.mod
    cp -r .just/templates/tests/main.go {{module}}/tests/main.go
    cd {{module}}/tests && dagger install ..
    dagger develop -m {{module}}/tests

    @echo ""
    @echo "Module \"{{module}}\" initialized"
    @echo "Don't forget to add it to GitHub Actions when you are ready!"

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

# run tests for a module
[group('dev')]
test module test='all':
    dagger call -m {{module}}/tests {{env('ARGS', '')}} {{test}}

# run examples for a module
[group('dev')]
examples module example='all':
    dagger call -m {{module}}/examples/go {{env('ARGS', '')}} {{example}}

# run `dagger develop` for all modules
[group('dev')]
develop:
    for dir in $(just list); do dagger develop -m $dir; done
    for dir in $(just list-with-tests); do dagger develop -m "$dir/tests"; done
    for dir in $(just list-with-examples); do dagger develop -m "$dir/examples/go"; done

# run `go mod tidy` for all modules
[group('dev')]
tidy:
    for dir in $(just list); do $(cd $dir; go mod tidy); done
    for dir in $(just list-with-tests); do $(cd "$dir/tests"; go mod tidy); done
    for dir in $(just list-with-examples); do $(cd "$dir/examples/go"; go mod tidy); done

# list modules (directories with a `dagger.json` file)
[group('list')]
@list:
    fd --type f --glob 'dagger.json' --exclude '**/.exclude' --exact-depth 2 --exec-batch dirname | xargs -I{} basename {} | sort -u

# list modules that have tests
[group('list')]
@list-with-tests:
    just list | while read -r dir; do if [ -d "$dir/tests" ]; then echo "$dir"; fi done

# list modules that have examples
[group('list')]
@list-with-examples:
    just list | while read -r dir; do if [ -d "$dir/examples/go" ]; then echo "$dir"; fi done

[private]
@as-json:
    jq -R -s -c 'split("\n") | map(select(length > 0))'
