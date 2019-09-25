{ pkgs ? import <nixpkgs> {} }:
with pkgs;

assert lib.versionAtLeast go.version "1.11";
buildGoPackage rec {
  name = "gogo-${version}";
  version = "0.0.1";
  goPackagePath = "github.com/PsyanticY/gogo";

  goDeps = ./deps.nix;
  src = ./.;
  #allowGoReference = true;

  meta = with stdenv.lib; {
    description = "simple go cli";
    homepage = https://github.com/PsyanticY/gogo;
    license = licenses.mit;
  };

}
