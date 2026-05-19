const { execSync } = require("child_process");
const { readdirSync, statSync, mkdirSync, existsSync, rmSync } = require("fs");
const { join } = require("path");

const SCHEMA_DIR = join(__dirname, "../../schema");
const TS_OUT_DIR = join(__dirname, "../src/types");

function getDirectories(srcPath) {
  return readdirSync(srcPath).filter(function(file) {
    return statSync(join(srcPath, file)).isDirectory();
  });
}

function getSchemaFiles(domainDir) {
  return readdirSync(domainDir)
    .filter(function(file) { return file.endsWith(".schema") || file.includes(".schema"); })
    .map(function(file) { return join(domainDir, file); });
}

function ensureDir(dir) {
  if (!existsSync(dir)) {
    mkdirSync(dir, { recursive: true });
  }
}

function generateTypeScript(domain, schemaFiles) {
  if (schemaFiles.length === 0) return;
  const tsDomainDir = join(TS_OUT_DIR, domain);
  ensureDir(tsDomainDir);
  const tsIndexFile = join(tsDomainDir, "index.ts");

  // 스키마 간 $ref 참조를 해결하기 위해 작업 디렉토리를 스키마 디렉토리로 변경
  const schemaDomainDir = join(SCHEMA_DIR, domain);
  const relativeSchemaFiles = schemaFiles.map(f => require("path").basename(f));
  const absoluteOutputFile = require("path").resolve(tsIndexFile);

  // npx quicktype 사용 (크로스 플랫폼 호환)
  const args = [
    "quicktype",
    "--src-lang",
    "schema",
    ...relativeSchemaFiles.map(f => `"${f}"`),
    "--lang",
    "typescript-zod",
    "--out",
    `"${absoluteOutputFile}"`
  ];

  const command = `npx ${args.join(" ")}`;

  try {
    execSync(command, {
      stdio: "inherit",
      shell: true,
      cwd: schemaDomainDir,  // 스키마 디렉토리에서 실행하여 $ref 참조 해결
      maxBuffer: 10 * 1024 * 1024
    });
    console.log("Generated TypeScript types for domain: " + domain);
  } catch (error) {
    console.warn("Warning: Failed to generate types for domain: " + domain);
    console.warn("  This may be due to schema validation issues or missing type names.");
  }
}

function main() {
  // 기존 타입 디렉토리 삭제 (스키마 변경사항 완전 반영)
  if (existsSync(TS_OUT_DIR)) {
    console.log("Cleaning existing types directory...");
    rmSync(TS_OUT_DIR, { recursive: true, force: true });
  }
  
  // 타입 디렉토리 재생성
  mkdirSync(TS_OUT_DIR, { recursive: true });
  
  var domains = getDirectories(SCHEMA_DIR);
  for (var i = 0; i < domains.length; i++) {
    var domain = domains[i];
    var domainDir = join(SCHEMA_DIR, domain);
    var schemaFiles = getSchemaFiles(domainDir);
    generateTypeScript(domain, schemaFiles);
  }
  console.log("TypeScript schema generation complete!");
}

main();
