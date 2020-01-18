{ system ? builtins.currentSystem }:

let
  pkgs = import <nixpkgs> { inherit system; };
in
  with pkgs; stdenv.mkDerivation rec {
    name    = "elmobd-${version}";
    version = builtins.replaceStrings ["\n"] [""] (builtins.readFile ./VERSION);

    # go get golang.org/x/tools/cmd/cover

    buildInputs = [
      go
      golint
      python3
    ];
  }
