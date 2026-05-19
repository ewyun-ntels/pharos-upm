import type {Role} from '@pharos/shared/types/user';

const GROUP_ORDER: Record<string, number> = {
  composite: 0,
  role: 2,
  attribute: 3,
};

function getGroupOrder(group?: string | null): number {
  if (!group) return 2;
  if (GROUP_ORDER[group] !== undefined) return GROUP_ORDER[group];
  return 1; // custom 그룹 (알파벳순으로 묶임)
}

export function sortRoles(roles: Role[]): Role[] {
  return [...roles].sort((a, b) => {
    const ga = a.group ?? null;
    const gb = b.group ?? null;
    const oa = getGroupOrder(ga);
    const ob = getGroupOrder(gb);
    if (oa !== ob) return oa - ob;
    if (ga !== gb) return (ga ?? '').localeCompare(gb ?? '');
    return a.display_name.localeCompare(b.display_name);
  });
}

export function getRoleBadgeDisplay(role: Role): {isAttr: boolean; prefix: string | null; displayName: string} {
  const isAttr = role.role.startsWith('attr:');
  const hasGroup = !!role.group;
  const prefix = hasGroup ? role.group! : role.role.substring(0, role.role.indexOf(':')) || null;
  const displayName =
    !hasGroup && prefix && role.display_name.startsWith(`${prefix}:`)
      ? role.display_name.substring(prefix.length + 1)
      : role.display_name;
  return {isAttr, prefix, displayName};
}
