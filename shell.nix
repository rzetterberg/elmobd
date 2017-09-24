{ system ? builtins.currentSystem }:

let
  pkgs = import <nixpkgs> { inherit system; };
in
  with pkgs; stdenv.mkDerivation rec {
    name    = "elmobd-${version}";
    version = "0.1.0";

    buildInputs = [
      go
    ];
  }
