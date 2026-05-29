#!/usr/bin/env tsx

import * as fs from 'fs';
import * as path from 'path';

// ─── Types ────────────────────────────────────────────────────────────────────

interface MenuOrderItem {
  extension: string;
  name: string;
  display_name?: string;
  hidden?: boolean;
  flat?: boolean;
  items?: MenuOrderItem[];
}

interface MenuOrderGroup {
  label?: string;
  separator?: boolean;
  hidden?: boolean;
  items: MenuOrderItem[];
}

interface SiteModeMeta {
  title: string;
  description?: string;
  icon?: string;
  'login-extension'?: string;
  'home-extension'?: string;
  'home-path'?: string;
  extensions?: string[];
  'menu-order'?: MenuOrderGroup[];
}

interface ExtensionEntry {
  name: string;
  dir: string;
}

export interface ViteExtensionEntry {
  name: string;
  path: string;
}

export interface ViteExtensions {
  extensions: ViteExtensionEntry[];
}

// ─── Utilities ────────────────────────────────────────────────────────────────

function findProjectRoot(): string {
  let dir = process.cwd();
  while (dir !== path.dirname(dir)) {
    if (fs.existsSync(path.join(dir, 'meta', 'configs'))) return dir;
    dir = path.dirname(dir);
  }
  console.error('Could not find meta/configs directory.');
  process.exit(1);
}

function loadSiteConfig(projectRoot: string, mode: string): SiteModeMeta {
  const configPath = path.join(projectRoot, 'meta', 'configs', `${mode}.json`);
  if (!fs.existsSync(configPath)) {
    console.error(`✗ Error: No config file found for SITE_MODE "${mode}"`);
    console.error(`  Expected: meta/configs/${mode}.json`);
    process.exit(1);
  }
  try {
    return JSON.parse(fs.readFileSync(configPath, 'utf-8'));
  } catch (e) {
    console.error(`Failed to load meta/configs/${mode}.json:`, e);
    process.exit(1);
  }
}

function scanExtensions(extensionsDir: string, targetExtensions: string[]): ExtensionEntry[] {
  const results: ExtensionEntry[] = [];
  const matchAll = targetExtensions.includes('all');

  function scan(dir: string, relativePath: string) {
    for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
      if (!entry.isDirectory() || entry.name === 'node_modules') continue;

      const currentPath = path.join(dir, entry.name);
      const currentRelative = relativePath ? `${relativePath}/${entry.name}` : entry.name;
      const packageJsonPath = path.join(currentPath, 'frontend', 'package.json');

      if (fs.existsSync(packageJsonPath)) {
        if (matchAll || targetExtensions.includes(currentRelative)) {
          // Generate expected package name from directory path
          const expectedName = `@pharos/extension-${currentRelative.replace(/\//g, '-')}`;

          // Read current package.json
          const pkg = JSON.parse(fs.readFileSync(packageJsonPath, 'utf-8'));

          // Auto-fix if mismatch
          if (pkg.name !== expectedName) {
            console.log(`✓ Auto-fixed package name: ${currentRelative}`);
            console.log(`  Old: ${pkg.name}`);
            console.log(`  New: ${expectedName}`);
            pkg.name = expectedName;
            fs.writeFileSync(packageJsonPath, JSON.stringify(pkg, null, 2) + '\n');
          }

          results.push({ name: expectedName, dir: currentRelative });
        }
      }

      scan(currentPath, currentRelative);
    }
  }

  scan(extensionsDir, '');
  return results;
}

function flattenMenuItems(items: MenuOrderItem[]): { extension: string; name: string }[] {
  const result: { extension: string; name: string }[] = [];
  for (const item of items) {
    if (!item.extension) {
      console.error(`✗ Error: menu-order item "${item.name}" is missing required "extension" field`);
      process.exit(1);
    }
    if (!item.name) {
      console.error(`✗ Error: menu-order item is missing required "name" field`);
      process.exit(1);
    }
    result.push({ extension: item.extension, name: item.name });
    if (item.items?.length) result.push(...flattenMenuItems(item.items));
  }
  return result;
}

const MENU_ORDER_INTERFACES = `
export interface MenuOrderItem {
  extension: string;
  name: string;
  display_name?: string;
  hidden?: boolean;
  flat?: boolean;
  items?: MenuOrderItem[];
}

export interface MenuOrderGroup {
  label?: string;
  separator?: boolean;
  hidden?: boolean;
  items: MenuOrderItem[];
}
`.trimStart();

// ─── Generators ───────────────────────────────────────────────────────────────

function validateLoginExtension(projectRoot: string, meta: SiteModeMeta) {
  const loginExtension = meta['login-extension'];
  if (!loginExtension) return;

  const extensionPath = path.join(projectRoot, 'extensions', loginExtension, 'frontend', 'package.json');
  if (!fs.existsSync(extensionPath)) {
    console.error(`✗ Error: Login extension "${loginExtension}" not found`);
    console.error(`  Expected: extensions/${loginExtension}/frontend/package.json`);
    process.exit(1);
  }
  console.log(`✓ Validated login-extension: ${loginExtension}`);
}

function validateHomeExtension(projectRoot: string, meta: SiteModeMeta) {
  const homeExtension = meta['home-extension'];
  if (!homeExtension) return;

  const configuredExtensions = meta.extensions ?? [];
  if (!configuredExtensions.includes('all') && !configuredExtensions.includes(homeExtension)) {
    console.error(`✗ Error: home-extension "${homeExtension}" must also be listed in extensions`);
    console.error(`  Example: "extensions": [..., "${homeExtension}"]`);
    process.exit(1);
  }

  const extensionPath = path.join(projectRoot, 'extensions', homeExtension, 'frontend', 'package.json');
  if (!fs.existsSync(extensionPath)) {
    console.error(`✗ Error: Home extension "${homeExtension}" not found`);
    console.error(`  Expected: extensions/${homeExtension}/frontend/package.json`);
    process.exit(1);
  }

  console.log(`✓ Validated home-extension: ${homeExtension}`);
}

function generateEnvFile(projectRoot: string, meta: SiteModeMeta) {
  const icon = meta.icon ? `/ui/images/${meta.icon}` : '';
  fs.writeFileSync(
    path.join(projectRoot, 'core', 'frontend', '.env.local'),
    `VITE_TITLE=${meta.title}\nVITE_ICON=${icon}\n`,
  );
  console.log('✓ Generated .env.local');
}

function copyFavicon(projectRoot: string, meta: SiteModeMeta) {
  if (!meta.icon) return;

  const iconRoot = meta.icon.split('/')[0];
  const srcDir = path.join(projectRoot, 'extensions', iconRoot, 'images');
  if (!fs.existsSync(srcDir)) return;

  const destDir = path.join(projectRoot, 'core', 'frontend', 'public', 'images', iconRoot, 'images');
  fs.mkdirSync(destDir, { recursive: true });

  let copied = 0;
  for (const file of fs.readdirSync(srcDir)) {
    fs.copyFileSync(path.join(srcDir, file), path.join(destDir, file));
    copied++;
  }
  if (copied > 0) console.log(`✓ Copied ${copied} image(s) from extensions/${iconRoot}/images`);
}

function generateSiteConfig(projectRoot: string, meta: SiteModeMeta) {
  const homeExtension = meta['home-extension']?.trim();
  const homePath = meta['home-path']?.trim() || (homeExtension ? `/extensions/${homeExtension}` : '/home');
  if (!homePath.startsWith('/')) {
    console.error(`✗ Error: home-path must start with "/"`);
    console.error(`  Received: ${homePath}`);
    process.exit(1);
  }

  const generatedDir = path.join(projectRoot, 'meta', 'generated', 'frontend');
  fs.writeFileSync(
    path.join(generatedDir, 'site-config.ts'),
    `/**\n * Site Config (AUTO-GENERATED)\n * Generated by: pnpm generate:metadata\n */\n\nexport const HOME_PATH = ${JSON.stringify(homePath)};\n`,
  );
  console.log(`✓ Generated site-config.ts (HOME_PATH: ${homePath})`);
}

function generateExtensionLoader(
  projectRoot: string,
  meta: SiteModeMeta,
  mode: string,
): ViteExtensions {
  const generatedDir = path.join(projectRoot, 'meta', 'generated', 'frontend');
  const extensionsDir = path.join(projectRoot, 'extensions');

  const targetExtensions = [
    ...(meta['login-extension'] ? [meta['login-extension']] : []),
    ...(meta.extensions ?? []),
  ];

  const extensions = scanExtensions(extensionsDir, targetExtensions);

  const imports = extensions
    .map(ext => `// Extension: ${ext.dir}\nimport '${ext.name}';`)
    .join('\n') || '// No extensions found';

  fs.writeFileSync(
    path.join(generatedDir, 'extension-loader.ts'),
    `/**\n * Extension Loader (AUTO-GENERATED)\n * Generated by: pnpm generate:metadata\n */\n\n// ========================================\n// AUTO-GENERATED IMPORTS\n// ========================================\n${imports}\n// ========================================\n// END AUTO-GENERATED\n// ========================================\n`,
  );
  console.log(`✓ Generated extension-loader.ts (${extensions.length} extension${extensions.length !== 1 ? 's' : ''} for SITE_MODE: ${mode})`);

  const viteExtensions: ViteExtensions = {
    extensions: extensions.map(ext => ({
      name: ext.name,
      path: `../../extensions/${ext.dir}/frontend/src`,
    })),
  };

  fs.writeFileSync(
    path.join(generatedDir, 'vite-extensions.json'),
    JSON.stringify(viteExtensions, null, 2) + '\n',
  );
  console.log(`✓ Generated vite-extensions.json (${extensions.length} extensions)`);

  return viteExtensions;
}

function generateMenuOrder(projectRoot: string, meta: SiteModeMeta, mode: string) {
  const generatedDir = path.join(projectRoot, 'meta', 'generated', 'frontend');
  const menuOrder = meta['menu-order'];
  const outputPath = path.join(generatedDir, 'menu-order.ts');

  if (!menuOrder?.length) {
    console.warn(`⚠ No menu-order defined for mode "${mode}". Sidebar order will be undefined.`);
    fs.writeFileSync(
      outputPath,
      `/**\n * Menu Order (AUTO-GENERATED)\n * Generated by: pnpm generate:metadata\n */\n\n${MENU_ORDER_INTERFACES}\nexport const MENU_ORDER: MenuOrderGroup[] = [];\n`,
    );
    console.log('✓ Generated menu-order.ts (empty)');
    return;
  }

  const allItems = menuOrder.flatMap(group => {
    if (!Array.isArray(group.items)) {
      console.error(`✗ Error: menu-order group is missing required "items" array`);
      process.exit(1);
    }
    return flattenMenuItems(group.items);
  });

  const keys = allItems.map(i => `${i.extension}:${i.name}`);
  const duplicates = keys.filter((k, i) => keys.indexOf(k) !== i);
  if (duplicates.length > 0) {
    console.error(`✗ Error: Duplicate menu keys in menu-order: ${Array.from(new Set(duplicates)).join(', ')}`);
    process.exit(1);
  }
  console.log(`✓ Validated menu-order: ${allItems.length} items, no duplicates`);

  fs.writeFileSync(
    outputPath,
    `/**\n * Menu Order (AUTO-GENERATED)\n * Generated by: pnpm generate:metadata\n * Source: meta/configs/${mode}.json → ['menu-order']\n *\n * DO NOT EDIT MANUALLY.\n */\n\n${MENU_ORDER_INTERFACES}\nexport const MENU_ORDER: MenuOrderGroup[] = ${JSON.stringify(menuOrder, null, 2)};\n`,
  );
  console.log(`✓ Generated menu-order.ts (${allItems.length} items for SITE_MODE: ${mode})`);
}

// ─── Main ─────────────────────────────────────────────────────────────────────

function main() {
  const mode = process.env.SITE_MODE?.trim() || 'default';
  console.log('Generating metadata for mode:', mode);

  const projectRoot = findProjectRoot();
  const meta = loadSiteConfig(projectRoot, mode);

  fs.mkdirSync(path.join(projectRoot, 'meta', 'generated', 'frontend'), { recursive: true });

  validateLoginExtension(projectRoot, meta);
  validateHomeExtension(projectRoot, meta);
  generateEnvFile(projectRoot, meta);
  copyFavicon(projectRoot, meta);
  generateSiteConfig(projectRoot, meta);
  generateExtensionLoader(projectRoot, meta, mode);
  generateMenuOrder(projectRoot, meta, mode);

  console.log('✓ All frontend metadata generated successfully');
}

main();
