#!/usr/bin/env node

const { execFileSync } = require("child_process");
const path = require("path");
const fs = require("fs");

function getPlatformPackage() {
  const platform = process.platform;
  const arch = process.arch;

  const packages = {
    "darwin-arm64": "@gitdig/darwin-arm64",
    "darwin-x64": "@gitdig/darwin-x64",
    "linux-arm64": "@gitdig/linux-arm64",
    "linux-x64": "@gitdig/linux-x64",
    "win32-arm64": "@gitdig/win32-arm64",
    "win32-x64": "@gitdig/win32-x64",
  };

  const pkg = packages[`${platform}-${arch}`];
  if (!pkg) {
    throw new Error(`Unsupported platform: ${platform}-${arch}`);
  }
  return pkg;
}

function getBinaryPath() {
  const localBin = path.join(__dirname, "bin", "gitdig");
  if (fs.existsSync(localBin)) {
    return localBin;
  }

  try {
    const pkgName = getPlatformPackage();
    const pkgPath = require.resolve(`${pkgName}/package.json`);
    const binaryName = process.platform === "win32" ? "gitdig.exe" : "gitdig";
    const binary = path.join(path.dirname(pkgPath), "bin", binaryName);
    if (fs.existsSync(binary)) {
      return binary;
    }
  } catch (e) {}

  throw new Error("GitDig binary not found. Try reinstalling: npm install -g gitdig");
}

function run() {
  const binary = getBinaryPath();
  try {
    execFileSync(binary, process.argv.slice(2), { stdio: "inherit" });
  } catch (error) {
    if (error.status !== undefined) {
      process.exit(error.status);
    }
    throw error;
  }
}

module.exports = { getBinaryPath, run };

if (require.main === module) {
  run();
}
