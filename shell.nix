{ system ? builtins.currentSystem }:

let
  pkgs = import <nixpkgs> { inherit system; };
in
  with pkgs; stdenv.mkDerivation rec {
    name    = "elmobd-${version}";
    version = builtins.replaceStrings ["\n"] [""] (builtins.readFile ./VERSION);

    buildInputs = [
      go
      python3
    ];
  }
