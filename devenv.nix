{ pkgs, inputs, ... }:

{
  cachix.pull = [ "sagikazarmark-dev" ];

  overlays = [
    (final: prev: {
      dagger = inputs.dagger.packages.${final.stdenv.hostPlatform.system}.dagger;
    })
  ];

  languages = {
    go = {
      enable = true;
      package = pkgs.go_1_25;
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
