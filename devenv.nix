{ pkgs, ... }:

{
  cachix.pull = [ "sagikazarmark-dev" ];

  languages = {
    go = {
      enable = true;
      package = pkgs.go_1_26;
    };
  };

  packages = with pkgs; [
    dagger
    golangci-lint
    just
    git
    semver-tool
    jq
    moreutils
    fd

    # is this still necessary?
    yq-go
  ];
}
