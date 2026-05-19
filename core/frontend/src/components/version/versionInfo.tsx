import React from 'react';
import {useOne} from '@/lib/data-provider';
import {Badge} from '@pharos/shared/components/ui';
import {UI_CONFIG_PROVIDER_NAME, UI_CONFIG_RESOURCES} from '@providers/ui-config-provider';

function VersionInfo() {
  const {query: {data: versionInfo}} = useOne({
    dataProviderName: UI_CONFIG_PROVIDER_NAME,
    resource: UI_CONFIG_RESOURCES.BADGES,
    id: 'app-version',
  });

  const version = versionInfo?.data?.value;

  if (!version) return null;

  return (
    <Badge variant="secondary" className="flex items-center space-x-1 shrink-0 mr-2 rounded-full">
      <p className="text-xs">
        <span className="text-muted-foreground text-[11px]">Version :</span> {version}
      </p>
    </Badge>
  );
}

export default VersionInfo;
