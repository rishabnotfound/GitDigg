#!/usr/bin/env node

const fs = require("fs");
const path = require("path");

const BINARY = process.platform === "win32" ? "gitdig.exe" : "gitdig";
const BIN_DIR = path.join(__dirname, "bin");

function tryPlatformPackage() {
  const packages = {
    "darwin-arm64": "@gitdig/darwin-arm64",
    "darwin-x64": "@gitdig/darwin-x64",
    "linux-arm64": "@gitdig/linux-arm64",
    "linux-x64": "@gitdig/linux-x64",
    "win32-arm64": "@gitdig/win32-arm64",
    "win32-x64": "@gitdig/win32-x64",
  };

  const pkg = packages[`${process.platform}-${process.arch}`];
  if (!pkg) {
    console.error(`Unsupported platform: ${process.platform}-${process.arch}`);
    process.exit(1);
  }

  try {
    const pkgPath = require.resolve(`${pkg}/package.json`);
    const binary = path.join(path.dirname(pkgPath), "bin", BINARY);

    if (fs.existsSync(binary)) {
      if (!fs.existsSync(BIN_DIR)) {
        fs.mkdirSync(BIN_DIR, { recursive: true });
      }

      const target = path.join(BIN_DIR, BINARY);
      if (fs.existsSync(target)) fs.unlinkSync(target);

      if (process.platform === "win32") {
        fs.writeFileSync(path.join(BIN_DIR, "gitdig.cmd"), `@echo off\n"${binary}" %*`);
      } else {
        fs.writeFileSync(target, `#!/bin/sh\nexec "${binary}" "$@"`);
        fs.chmodSync(target, 0o755);
      }

      console.log("GitDig ready");
      return true;
    }
  } catch (e) {}

  return false;
}

function install() {
  if (process.env.GITDIG_SKIP_INSTALL) {
    return;
  }

  if (!tryPlatformPackage()) {
    console.error("Platform package not found. Try reinstalling.");
    process.exit(1);
  }
}

install();
