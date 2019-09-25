{ pkgs ? import <nixpkgs> {} }:
with pkgs;

assert lib.versionAtLeast go.version "1.11";

buildGoPackage rec {
  name = "gogo-${version}";
  version = "0.0.1";
  goPackagePath = "github.com/PsyanticY/gogo";

  nativeBuildInputs = [ makeWrapper ];

  goDeps = ./deps.nix;
  src = ./.;

  postInstall = with stdenv; let
    binPath = lib.makeBinPath [ nix-prefetch-git go ];
  in ''
    wrapProgram $bin/bin/gogo --prefix PATH : ${binPath}
  '';

  meta = with stdenv.lib; {
    description = "simple go cli";
    homepage = https://github.com/PsyanticY/gogo;
    license = licenses.mit;
  };

}
