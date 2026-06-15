#!/usr/bin/env node

const assert = require("assert");
const crypto = require("crypto");
const { spawnSync } = require("child_process");
const fs = require("fs");
const os = require("os");
const path = require("path");

const {
  archiveNameForTarget,
  archiveURLForTarget,
  binaryNameForPlatform,
  checksumNameForVersion,
  checksumURLForVersion,
  createWindowsBpShim,
  normalizeBaseURL,
  parseChecksumFile,
  targetForPlatform,
  verifyChecksum,
  version,
} = require("./install");
const pkg = require("./package.json");

const packagedBp = fs.readFileSync(path.join(__dirname, "bin", "bp"), "utf8");
assert.ok(packagedBp.startsWith("#!/usr/bin/env node"));

function withTempDir(fn) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "bp-npm-test-"));
  try {
    fn(dir);
  } finally {
    fs.rmSync(dir, { recursive: true, force: true });
  }
}

assert.strictEqual(pkg.bin.bp, "bin/bp");
assert.strictEqual(pkg.name, "@byteplus/cli");
assert.strictEqual(version, pkg.version);
assert.strictEqual(pkg.repository.url, "https://github.com/byteplus-sdk/byteplus-cli");
assert.strictEqual(binaryNameForPlatform("win32"), "bp.exe");
assert.strictEqual(binaryNameForPlatform("linux"), "bp");
assert.strictEqual(binaryNameForPlatform("darwin"), "bp");
assert.strictEqual(normalizeBaseURL("https://example.com/base///"), "https://example.com/base");
assert.deepStrictEqual(targetForPlatform("darwin", "arm64"), {
  platform: "darwin",
  arch: "arm64",
});
assert.deepStrictEqual(targetForPlatform("linux", "arm"), {
  platform: "linux",
  arch: "arm",
});
assert.deepStrictEqual(targetForPlatform("win32", "arm64"), {
  platform: "windows",
  arch: "arm64",
});
assert.strictEqual(targetForPlatform("win32", "arm"), null);
assert.strictEqual(
  archiveNameForTarget({ platform: "darwin", arch: "arm64" }, "1.2.3"),
  "byteplus-cli_1.2.3_darwin_arm64.zip"
);
assert.strictEqual(
  archiveURLForTarget(
    { platform: "linux", arch: "amd64" },
    "1.2.3",
    "https://bucket.example.com/releases///"
  ),
  "https://bucket.example.com/releases/v1.2.3/byteplus-cli_1.2.3_linux_amd64.zip"
);
assert.strictEqual(checksumNameForVersion("1.2.3"), "byteplus-cli_1.2.3_SHA256SUMS");
assert.strictEqual(
  checksumURLForVersion("1.2.3", "https://bucket.example.com/releases///"),
  "https://bucket.example.com/releases/v1.2.3/byteplus-cli_1.2.3_SHA256SUMS"
);

const archiveName = archiveNameForTarget({ platform: "linux", arch: "amd64" }, version);
const archiveData = Buffer.from("fake archive");
const archiveHash = crypto.createHash("sha256").update(archiveData).digest("hex");
const checksumContent = [
  `${archiveHash}  ${archiveName}`,
  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  byteplus-cli_extra.zip",
  "",
].join("\n");
assert.deepStrictEqual(parseChecksumFile(checksumContent)[archiveName], archiveHash);
assert.doesNotThrow(() => verifyChecksum(archiveName, archiveData, checksumContent));
assert.throws(
  () => verifyChecksum(archiveName, Buffer.from("tampered"), checksumContent),
  /Checksum mismatch/
);
assert.throws(
  () => verifyChecksum("missing.zip", archiveData, checksumContent),
  /Checksum for missing\.zip not found/
);

withTempDir((dir) => {
  const binDir = path.join(dir, "bin");
  fs.mkdirSync(binDir, { recursive: true });

  const exePath = path.join(binDir, "bp.exe");
  fs.writeFileSync(exePath, "#!/bin/sh\necho bp.exe \"$@\"\nexit 7\n");
  fs.chmodSync(exePath, 0o755);

  createWindowsBpShim(binDir);

  const shimPath = path.join(binDir, "bp");
  assert.ok(fs.existsSync(shimPath), "Windows npm entry bin/bp should exist");

  if (process.platform === "win32") {
    const shim = fs.readFileSync(shimPath, "utf8");
    assert.ok(shim.includes('path.join(__dirname, "bp.exe")'));
  } else {
    const result = spawnSync(process.execPath, [shimPath, "arg1", "arg2"], {
      encoding: "utf8",
    });
    assert.strictEqual(result.status, 7);
    assert.strictEqual(result.stdout.trim(), "bp.exe arg1 arg2");
  }
});

console.log("install tests passed");
