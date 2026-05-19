import React from 'react';
import {Popover, PopoverContent, PopoverTrigger} from '@pharos/shared/components/ui';
import {Button} from '@pharos/shared/components/ui';
import {ChevronDown} from '@pharos/shared/components';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
} from '@pharos/shared/components/ui';
import {CommandList} from '@pharos/shared/components/ui-extension';
import * as yaml from 'js-yaml';
import {pod} from '@components/editor/templates/pod';
import {service} from '@components/editor/templates/service';
import {statefulSet} from '@components/editor/templates/statefulSet';
import {secret} from '@components/editor/templates/secret';
import {configMap} from '@components/editor/templates/configMap';
import {deployment} from '@components/editor/templates/deployment';
import {persistentVolume} from '@components/editor/templates/persistentVolume';
import {persistentVolumeClaim} from '@components/editor/templates/persistentVolumeClaim';
import {cronJob} from '@components/editor/templates/cronJob';
import {job} from '@components/editor/templates/job';
import {daemonset} from '@components/editor/templates/daemonset';
import {replicaSet} from '@components/editor/templates/replicaSet';
import {clusterRole} from '@components/editor/templates/clusterRole';
import {clusterRoleBinding} from '@components/editor/templates/clusterRoleBinding';
import {role} from '@components/editor/templates/role';
import {roleBinding} from '@components/editor/templates/roleBinding';

const templates = new Map([
  ['Pod', yaml.dump(pod)],
  ['Service', yaml.dump(service)],
  ['StatefulSet', yaml.dump(statefulSet)],
  ['Secret', yaml.dump(secret)],
  ['ConfigMap', yaml.dump(configMap)],
  ['Deployment', yaml.dump(deployment)],
  ['PersistentVolume', yaml.dump(persistentVolume)],
  ['PersistentVolumeClaim', yaml.dump(persistentVolumeClaim)],
  ['CronJob', yaml.dump(cronJob)],
  ['Job', yaml.dump(job)],
  ['DaemonSet', yaml.dump(daemonset)],
  ['ReplicaSet', yaml.dump(replicaSet)],
  ['ClusterRole', yaml.dump(clusterRole)],
  ['ClusterRoleBinding', yaml.dump(clusterRoleBinding)],
  ['Role', yaml.dump(role)],
  ['RoleBinding', yaml.dump(roleBinding)],
]);

interface TemplateProps {
  onSelect: (data: string | undefined) => void;
}

const TemplateSelector = ({onSelect}: TemplateProps) => {
  const [open, setOpen] = React.useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button variant="outline" className="h-8">
          Select Template ...{<ChevronDown className={'h-4 w-4'} />}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-52 h-92 p-1">
        <Command
          filter={(value, search) => (value.toLowerCase().includes(search.toLowerCase()) ? 1 : 0)}
        >
          <CommandInput placeholder="filter template" />
          <CommandList>
            <CommandEmpty>No results found.</CommandEmpty>
            <CommandGroup>
              {Array.from(templates.entries()).map(([key, value]) => (
                <CommandItem
                  className={'w-full'}
                  key={`editor-template-${key}`}
                  value={key}
                  onSelect={() => {
                    onSelect(value);
                    setOpen(false);
                  }}
                >
                  {key}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
};

export const Template = TemplateSelector;
