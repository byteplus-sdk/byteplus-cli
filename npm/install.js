#!/usr/bin/env node

const { execFileSync } = require("child_process");
const crypto = require("crypto");
const https = require("https");
const fs = require("fs");
const path = require("path");

const VERSION = require("./package.json").version;
const DEFAULT_DOWNLOAD_BASE_URL = "https://byteplus-cli.tos-ap-southeast-1.bytepluses.com/bp";
const DOWNLOAD_BASE_URL = normalizeBaseURL(
  process.env.BYTEPLUS_CLI_DOWNLOAD_BASE_URL || DEFAULT_DOWNLOAD_BASE_URL
);

const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
  freebsd: "freebsd",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
  ia32: "386",
  arm: "arm",
};

const SUPPORTED_TARGETS = {
  darwin: ["amd64", "arm64"],
  linux: ["amd64", "386", "arm", "arm64"],
  freebsd: ["amd64", "386", "arm", "arm64"],
  windows: ["amd64", "386", "arm64"],
};

function normalizeBaseURL(url) {
  return String(url || "").replace(/\/+$/, "");
}

function binaryNameForPlatform(platform) {
  return platform === "win32" ? "bp.exe" : "bp";
}

function targetForPlatform(platform, arch) {
  const targetPlatform = PLATFORM_MAP[platform];
  const targetArch = ARCH_MAP[arch];

  if (!targetPlatform || !targetArch) {
    return null;
  }

  const supportedArchs = SUPPORTED_TARGETS[targetPlatform] || [];
  if (supportedArchs.indexOf(targetArch) === -1) {
    return null;
  }

  return {
    platform: targetPlatform,
    arch: targetArch,
  };
}

function archiveNameForTarget(target, version) {
  return `byteplus-cli_${version}_${target.platform}_${target.arch}.zip`;
}

function archiveURLForTarget(target, version, downloadBaseURL) {
  const baseURL = normalizeBaseURL(downloadBaseURL);
  return `${baseURL}/v${version}/${archiveNameForTarget(target, version)}`;
}

function checksumNameForVersion(version) {
  return `byteplus-cli_${version}_SHA256SUMS`;
}

function checksumURLForVersion(version, downloadBaseURL) {
  const baseURL = normalizeBaseURL(downloadBaseURL);
  return `${baseURL}/v${version}/${checksumNameForVersion(version)}`;
}

function parseChecksumFile(content) {
  const checksums = {};
  String(content || "")
    .split(/\r?\n/)
    .forEach((rawLine) => {
      const line = rawLine.trim();
      if (!line) {
        return;
      }

      const match = line.match(/^([a-fA-F0-9]{64})\s+\*?(.+)$/);
      if (!match) {
        return;
      }

      checksums[path.basename(match[2].trim())] = match[1].toLowerCase();
    });
  return checksums;
}

function sha256(data) {
  return crypto.createHash("sha256").update(data).digest("hex");
}

function verifyChecksum(archiveName, data, checksumContent) {
  const checksums = parseChecksumFile(checksumContent);
  const expected = checksums[archiveName];

  if (!expected) {
    throw new Error(`Checksum for ${archiveName} not found in ${checksumNameForVersion(VERSION)}`);
  }

  const actual = sha256(data);
  if (actual !== expected) {
    throw new Error(`Checksum mismatch for ${archiveName}: expected ${expected}, got ${actual}`);
  }
}

function createWindowsBpShim(binDir) {
  const shimPath = path.join(binDir, "bp");
  const shim = `#!/usr/bin/env node

const { spawnSync } = require("child_process");
const path = require("path");

const exePath = path.join(__dirname, "bp.exe");
const result = spawnSync(exePath, process.argv.slice(2), { stdio: "inherit" });

if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}

process.exit(result.status === null ? 1 : result.status);
`;

  fs.writeFileSync(shimPath, shim);
  fs.chmodSync(shimPath, 0o755);
}

function download(url) {
  return new Promise((resolve, reject) => {
    const follow = (currentURL) => {
      https
        .get(currentURL, (res) => {
          if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
            res.resume();
            follow(new URL(res.headers.location, currentURL).toString());
            return;
          }
          if (res.statusCode !== 200) {
            reject(new Error(`Download failed: HTTP ${res.statusCode} for ${currentURL}`));
            return;
          }
          const chunks = [];
          res.on("data", (chunk) => chunks.push(chunk));
          res.on("end", () => resolve(Buffer.concat(chunks)));
          res.on("error", reject);
        })
        .on("error", reject);
    };
    follow(url);
  });
}

async function install() {
  const target = targetForPlatform(process.platform, process.arch);

  if (!target) {
    console.error(`Unsupported platform: ${process.platform} ${process.arch}`);
    process.exit(1);
  }

  const zipName = archiveNameForTarget(target, VERSION);
  const url = archiveURLForTarget(target, VERSION, DOWNLOAD_BASE_URL);
  const checksumURL = checksumURLForVersion(VERSION, DOWNLOAD_BASE_URL);
  const binDir = path.join(__dirname, "bin");
  const isWindows = process.platform === "win32";
  const binaryName = binaryNameForPlatform(process.platform);
  const binPath = path.join(binDir, binaryName);

  fs.mkdirSync(binDir, { recursive: true });
  console.log(`Downloading ${zipName}...`);

  const [data, checksumData] = await Promise.all([download(url), download(checksumURL)]);
  console.log(`Verifying ${zipName}...`);
  verifyChecksum(zipName, data, checksumData.toString("utf8"));

  const tmpDir = path.join(__dirname, ".tmp");
  const zipPath = path.join(tmpDir, zipName);

  fs.mkdirSync(tmpDir, { recursive: true });
  fs.writeFileSync(zipPath, data);

  try {
    if (isWindows) {
      execFileSync(
        "powershell",
        [
          "-NoProfile",
          "-Command",
          "Expand-Archive -LiteralPath $args[0] -DestinationPath $args[1] -Force",
          zipPath,
          tmpDir,
        ],
        { stdio: "pipe" }
      );
    } else {
      execFileSync("unzip", ["-o", "-q", zipPath, "-d", tmpDir], { stdio: "pipe" });
    }

    const extracted = fs.readdirSync(tmpDir);
    const bpBinary = extracted.find((f) => f === "bp" || f === "bp.exe");

    if (!bpBinary) {
      console.error("Could not find 'bp' binary in zip archive. Found:", extracted);
      process.exit(1);
    }

    fs.copyFileSync(path.join(tmpDir, bpBinary), binPath);

    if (isWindows) {
      createWindowsBpShim(binDir);
    } else {
      fs.chmodSync(binPath, 0o755);
    }

    if (process.platform === "darwin") {
      try {
        execFileSync("xattr", ["-d", "com.apple.quarantine", binPath], { stdio: "pipe" });
      } catch (_) {
        // Attribute may not exist, ignore.
      }
    }

    console.log(`BytePlus CLI v${VERSION} installed for ${target.platform}/${target.arch}`);
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

if (require.main === module) {
  install().catch((err) => {
    console.error("Installation failed:", err.message);
    process.exit(1);
  });
}

module.exports = {
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
  version: VERSION,
};
