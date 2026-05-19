import React, {useState, useMemo, useEffect} from 'react';
import {useNavigate, useLocation} from 'react-router-dom';
import {useList, useOne, useCreate, HttpError} from '@/lib/data-provider';
import {
  Button,
  Input,
  Label,
  Textarea,
  Switch,
  Checkbox,
  Badge,
  Separator,
} from '@pharos/shared/components/ui';
import {useToast} from '@hooks/use-toast';
import {ROLE_PROVIDER_NAME, ROLE_RESOURCES} from '@providers/role-provider';
import type {RoleMetadata, RoleGroup} from '@features/settings/roles/types';
import {PageHeader} from '@components/page-header';

const BUILTIN_GROUPS = ['composite', 'custom', 'role', 'attribute'] as const;

export default function CompositeRoleFormPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const {toast} = useToast();

  const editTarget: RoleMetadata | undefined = (location.state as {editTarget?: RoleMetadata} | null)
    ?.editTarget;
  const isEdit = !!editTarget;

  // Fetch base roles (for selection)
  const {
    query: {data: baseData, isLoading: isLoadingBase},
  } = useList<RoleGroup>({
    dataProviderName: ROLE_PROVIDER_NAME,
    resource: ROLE_RESOURCES.ALL_ROLES,
    queryOptions: {retry: false, staleTime: 0},
    meta: {includeHidden: true},
  });

  // Fetch current overlay (to update on save)
  const {
    query: {data: overlayData, isLoading: isLoadingOverlay},
  } = useOne<RoleMetadata[]>({
    dataProviderName: ROLE_PROVIDER_NAME,
    resource: ROLE_RESOURCES.CONFIG,
    id: 'config',
    queryOptions: {retry: false, staleTime: 0},
  });

  const {mutateAsync: saveConfig} = useCreate<RoleMetadata[], HttpError, {config: RoleMetadata[]}>();
  const [isSaving, setIsSaving] = useState(false);

  // Form state
  const [key, setKey] = useState('');
  const [displayName, setDisplayName] = useState('');
  const [description, setDescription] = useState('');
  const [groupName, setGroupName] = useState('composite');
  const [hide, setHide] = useState(false);
  const [selectedRoles, setSelectedRoles] = useState<Record<string, boolean>>({});
  const [search, setSearch] = useState('');

  const baseGroups: RoleGroup[] = useMemo(
    () => (baseData?.data as unknown as RoleGroup[]) ?? [],
    [baseData],
  );

  const currentOverlay: RoleMetadata[] = useMemo(() => {
    const d = overlayData?.data as unknown;
    return Array.isArray(d) ? (d as RoleMetadata[]) : [];
  }, [overlayData]);

  // Initialize form from editTarget
  useEffect(() => {
    if (editTarget) {
      setKey(editTarget.key);
      setDisplayName(editTarget.displayName);
      setDescription(editTarget.description ?? '');
      setGroupName(editTarget.group ?? 'composite');
      setHide(editTarget.hide ?? false);
      setSelectedRoles({...(editTarget.roles ?? {})});
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // run once on mount

  const filteredGroups = useMemo(() => {
    if (!search.trim()) return baseGroups;
    const q = search.toLowerCase();
    return baseGroups
      .map((group) => ({
        ...group,
        roles: group.roles.filter(
          (r) => r.key.toLowerCase().includes(q) || r.displayName.toLowerCase().includes(q),
        ),
      }))
      .filter((group) => group.roles.length > 0);
  }, [baseGroups, search]);

  const selectedCount = Object.values(selectedRoles).filter(Boolean).length;
  const isValid = key.trim().length > 0 && displayName.trim().length > 0 && selectedCount > 0;
  const isLoading = isLoadingBase || isLoadingOverlay;

  const getGroupCheckState = (group: RoleGroup): boolean | 'indeterminate' => {
    const total = group.roles.length;
    const checked = group.roles.filter((r) => selectedRoles[r.key]).length;
    if (checked === 0) return false;
    if (checked === total) return true;
    return 'indeterminate';
  };

  const handleGroupToggle = (group: RoleGroup, currentState: boolean | 'indeterminate') => {
    const shouldSelect = currentState !== true;
    setSelectedRoles((prev) => {
      const next = {...prev};
      group.roles.forEach((r) => {
        next[r.key] = shouldSelect;
      });
      return next;
    });
  };

  const handleRoleToggle = (roleKey: string, checked: boolean) => {
    setSelectedRoles((prev) => ({...prev, [roleKey]: checked}));
  };

  const handleSave = async () => {
    if (!isValid) return;

    const rolesMap: Record<string, boolean> = {};
    Object.entries(selectedRoles).forEach(([k, v]) => {
      if (v) rolesMap[k] = true;
    });

    const newRole: RoleMetadata = {
      key: key.trim(),
      displayName: displayName.trim(),
      description: description.trim(),
      group: groupName.trim() || 'composite',
      hide,
      roles: rolesMap,
    };

    const existingComposites = currentOverlay.filter(
      (r) => r.roles && Object.keys(r.roles).length > 0,
    );
    const atomicOverlay = currentOverlay.filter(
      (r) => !r.roles || Object.keys(r.roles).length === 0,
    );

    const updatedComposites = (() => {
      const idx = existingComposites.findIndex((r) => r.key === newRole.key);
      if (idx >= 0) {
        const updated = [...existingComposites];
        updated[idx] = newRole;
        return updated;
      }
      return [...existingComposites, newRole];
    })();

    const configToSave: RoleMetadata[] = [...atomicOverlay, ...updatedComposites];

    setIsSaving(true);
    try {
      await saveConfig({
        resource: ROLE_RESOURCES.CONFIG,
        values: {config: configToSave},
        dataProviderName: ROLE_PROVIDER_NAME,
      });
      toast({description: 'Saved successfully.'});
      navigate('/settings/roles', {state: {activeTab: 'Composite Roles'}});
    } catch {
      toast({description: 'Failed to save.', variant: 'destructive'});
    } finally {
      setIsSaving(false);
    }
  };

  const handleCancel = () => {
    navigate('/settings/roles', {state: {activeTab: 'Composite Roles'}});
  };

  const title = isEdit ? 'Edit Composite Role' : 'Add Composite Role';

  return (
    <main className="flex flex-col w-full h-full overflow-hidden">
      <PageHeader
        title={title}
        breadcrumbLabel={title}
        actions={
          isEdit ? (
            <code className="text-xs font-mono bg-muted px-2 py-0.5 rounded text-muted-foreground">
              {editTarget.key}
            </code>
          ) : undefined
        }
      />

      {/* Two-column layout */}
      <div className="flex flex-1 overflow-hidden min-h-0">
        {/* Left panel: Role metadata */}
        <div className="w-80 shrink-0 border-r flex flex-col overflow-y-auto px-5 py-5 gap-5">
          <div className="space-y-1.5">
            <Label htmlFor="cr-key" className="text-sm">
              Key <span className="text-destructive">*</span>
            </Label>
            <Input
              id="cr-key"
              value={key}
              onChange={(e) => setKey(e.target.value)}
              disabled={isEdit}
              placeholder="role:custom_operator"
              className={`font-mono text-sm ${isEdit ? 'bg-muted cursor-not-allowed' : ''}`}
            />
            {!isEdit && (
              <p className="text-xs text-muted-foreground">
                Convention:{' '}
                <code className="font-mono bg-muted px-1 rounded">role:your_name</code>
              </p>
            )}
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="cr-name" className="text-sm">
              Display Name <span className="text-destructive">*</span>
            </Label>
            <Input
              id="cr-name"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="Custom Operator"
              className="text-sm"
            />
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="cr-desc" className="text-sm">
              Description
            </Label>
            <Textarea
              id="cr-desc"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Describe what this role grants..."
              rows={3}
              className="resize-none text-sm"
            />
          </div>

          <div className="flex items-center justify-between">
            <Label className="text-sm">Visible in UI</Label>
            <Switch checked={!hide} onCheckedChange={(v) => setHide(!v)} />
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="cr-group" className="text-sm">
              Group
            </Label>
            <Input
              id="cr-group"
              value={groupName}
              onChange={(e) => setGroupName(e.target.value)}
              placeholder="composite"
              className="font-mono text-sm"
            />
            <div className="flex flex-wrap gap-1.5">
              {BUILTIN_GROUPS.map((g) => (
                <button
                  key={g}
                  type="button"
                  onClick={() => setGroupName(g)}
                  className={`font-mono text-xs px-2 py-0.5 rounded border transition-colors ${
                    groupName === g
                      ? 'bg-primary text-primary-foreground border-primary'
                      : 'border-border text-muted-foreground hover:text-foreground hover:border-foreground/40'
                  }`}
                >
                  {g}
                </button>
              ))}
            </div>
            <p className="text-xs text-muted-foreground">
              Determines which card this role appears under in the user permission editor.
            </p>
          </div>

          {/* Action buttons at bottom */}
          <div className="mt-auto pt-2 flex gap-2">
            <Button
              variant="outline"
              onClick={handleCancel}
              className="flex-1"
              disabled={isSaving}
            >
              Cancel
            </Button>
            <Button
              onClick={handleSave}
              disabled={!isValid || isSaving || isLoading}
              className="flex-1"
            >
              {isSaving ? 'Saving…' : isEdit ? 'Update' : 'Create'}
            </Button>
          </div>
        </div>

        {/* Right panel: Base roles selector */}
        <div className="flex-1 flex flex-col overflow-hidden px-5 py-5 min-w-0">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <Label className="text-sm font-semibold">
                Base Roles <span className="text-destructive">*</span>
              </Label>
              <p className="text-xs text-muted-foreground">
                Select the atomic roles bundled into this composite role.
              </p>
            </div>
            {selectedCount > 0 && (
              <Badge variant="default" className="text-xs shrink-0">
                {selectedCount} selected
              </Badge>
            )}
          </div>

          <Input
            placeholder="Search roles…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="h-8 text-sm mb-3 shrink-0"
          />

          <div className="flex-1 border rounded-md overflow-y-auto min-h-0">
            {isLoading ? (
              <div className="px-3 py-6 text-sm text-muted-foreground text-center">
                Loading roles…
              </div>
            ) : filteredGroups.length === 0 ? (
              <div className="px-3 py-6 text-sm text-muted-foreground text-center">
                No roles found
              </div>
            ) : (
              filteredGroups.map((group, gi) => {
                const groupState = getGroupCheckState(group);
                const groupSelectedCount = group.roles.filter((r) => selectedRoles[r.key]).length;

                return (
                  <div key={group.group}>
                    {gi > 0 && <Separator />}

                    {/* Group header */}
                    <label className="flex items-center gap-3 px-3 py-2 bg-muted/40 cursor-pointer hover:bg-muted/70 transition-colors">
                      <Checkbox
                        checked={groupState}
                        onCheckedChange={() => handleGroupToggle(group, groupState)}
                        className="shrink-0"
                      />
                      <div className="flex items-center gap-2 flex-1 min-w-0">
                        <span className="text-sm font-semibold">{group.displayName}</span>
                        <Badge variant="secondary" className="font-mono text-xs shrink-0">
                          {group.group}
                        </Badge>
                        <span className="text-xs text-muted-foreground ml-auto shrink-0">
                          {groupSelectedCount}/{group.roles.length}
                        </span>
                      </div>
                    </label>

                    {/* Roles in group */}
                    <div className="divide-y divide-border/60">
                      {group.roles.map((role) => (
                        <label
                          key={role.key}
                          className="flex items-start gap-3 pl-9 pr-3 py-2.5 cursor-pointer hover:bg-muted/40 transition-colors"
                        >
                          <Checkbox
                            checked={!!selectedRoles[role.key]}
                            onCheckedChange={(checked) => handleRoleToggle(role.key, !!checked)}
                            className="mt-0.5 shrink-0"
                          />
                          <div className="min-w-0">
                            <div className="flex items-center gap-1.5 flex-wrap">
                              <span className="text-sm font-medium">{role.displayName}</span>
                              <Badge variant="outline" className="font-mono text-xs shrink-0">
                                {role.key}
                              </Badge>
                            </div>
                            {role.description && (
                              <p className="text-xs text-muted-foreground mt-0.5 leading-relaxed">
                                {role.description}
                              </p>
                            )}
                          </div>
                        </label>
                      ))}
                    </div>
                  </div>
                );
              })
            )}
          </div>

          {selectedCount === 0 && !isLoading && (
            <p className="text-xs text-destructive mt-2">
              At least one base role must be selected.
            </p>
          )}
        </div>
      </div>
    </main>
  );
}
