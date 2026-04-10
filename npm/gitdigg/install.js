#!/usr/bin/env node

const fs = require("fs");
const path = require("path");

const BINARY = process.platform === "win32" ? "gitdigg.exe" : "gitdigg";
const BIN_DIR = path.join(__dirname, "bin");

function tryPlatformPackage() {
  const packages = {
    "darwin-arm64": "@gitdigg/darwin-arm64",
    "darwin-x64": "@gitdigg/darwin-x64",
    "linux-arm64": "@gitdigg/linux-arm64",
    "linux-x64": "@gitdigg/linux-x64",
    "win32-arm64": "@gitdigg/win32-arm64",
    "win32-x64": "@gitdigg/win32-x64",
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
        fs.writeFileSync(path.join(BIN_DIR, "gitdigg.cmd"), `@echo off\n"${binary}" %*`);
      } else {
        fs.writeFileSync(target, `#!/bin/sh\nexec "${binary}" "$@"`);
        fs.chmodSync(target, 0o755);
      }

      console.log("GitDigg ready");
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
