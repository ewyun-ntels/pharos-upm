/**
 * SelectBox & MultiSelect Components Demo
 * 
 * Demonstrates the usage of SelectBox and MultiSelect components
 * that were refactored to pure Controlled Components.
 */

import React, { useState } from 'react';
import { SelectBox, MultiSelect } from '@pharos/shared/components';
import type { SelectOption } from '@pharos/shared/components';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@pharos/shared/components/ui';
import { Badge } from '@pharos/shared/components/ui';

// Sample data
const frameworks: SelectOption[] = [
  { label: 'React', value: 'react', icon: '⚛️' },
  { label: 'Vue', value: 'vue', icon: '💚' },
  { label: 'Angular', value: 'angular', icon: '🅰️' },
  { label: 'Svelte', value: 'svelte', icon: '🔥' },
  { label: 'Next.js', value: 'nextjs', icon: '▲' },
];

const languages = ['TypeScript', 'JavaScript', 'Python', 'Go', 'Rust', 'Java'];

const statusOptions = [
  { label: 'Active', value: 'active' },
  { label: 'Inactive', value: 'inactive' },
  { label: 'Pending', value: 'pending' },
  { label: 'Archived', value: 'archived' },
];

const teamMembers = [
  { label: 'Alice Johnson', value: 'alice' },
  { label: 'Bob Smith', value: 'bob' },
  { label: 'Charlie Brown', value: 'charlie' },
  { label: 'Diana Prince', value: 'diana' },
  { label: 'Ethan Hunt', value: 'ethan' },
  { label: 'Fiona Apple', value: 'fiona' },
];

export const SelectComponentsDemo: React.FC = () => {
  // SelectBox states
  const [selectedFramework, setSelectedFramework] = useState<string>('');
  const [selectedLanguage, setSelectedLanguage] = useState<string>('');
  const [selectedStatus, setSelectedStatus] = useState<string>('');

  // MultiSelect states
  const [selectedMembers, setSelectedMembers] = useState<string[]>([]);
  const [selectedLanguages, setSelectedLanguages] = useState<string[]>([]);

  return (
    <div className="fixed inset-0 overflow-y-auto bg-background" style={{ top: '64px' }}>
      <div className="container mx-auto p-6 space-y-6 min-h-full">
        <div className="space-y-2">
          <h1 className="text-3xl font-bold">SelectBox & MultiSelect Demo</h1>
          <p className="text-muted-foreground">
            Pure Controlled Components - No internal API calls, fully reusable
          </p>
        </div>

      {/* SelectBox Examples */}
      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>SelectBox - With Icons</CardTitle>
            <CardDescription>
              Select a framework from the list with custom icons
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <SelectBox
              label="Framework"
              options={frameworks}
              value={selectedFramework}
              onChange={setSelectedFramework}
              placeholder="Choose a framework..."
              size="medium"
            />
            <div className="text-sm">
              Selected: <Badge variant="secondary">{selectedFramework || 'None'}</Badge>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>SelectBox - String Array</CardTitle>
            <CardDescription>
              Simple string array options without auto-select
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <SelectBox
              label="Language"
              options={languages}
              value={selectedLanguage}
              onChange={setSelectedLanguage}
              placeholder="Select a language..."
              autoSelectFirstOption={false}
              size="medium"
            />
            <div className="text-sm">
              Selected: <Badge variant="secondary">{selectedLanguage || 'None'}</Badge>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>SelectBox - Small Size</CardTitle>
            <CardDescription>
              Compact select with auto-select enabled
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <SelectBox
              label="Status"
              options={statusOptions}
              value={selectedStatus}
              onChange={setSelectedStatus}
              size="small"
              autoSelectFirstOption={true}
            />
            <div className="text-sm">
              Selected: <Badge variant="secondary">{selectedStatus || 'None'}</Badge>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>SelectBox - Full Width</CardTitle>
            <CardDescription>
              Full width select with custom label key
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <SelectBox
              label="Project Status"
              options={statusOptions}
              value={selectedStatus}
              onChange={setSelectedStatus}
              size="full"
              labelKey="label"
              valueKey="value"
            />
          </CardContent>
        </Card>
      </div>

      {/* MultiSelect Examples */}
      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>MultiSelect - Team Members</CardTitle>
            <CardDescription>
              Select multiple team members with "Select All" option
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <MultiSelect
              label="Team"
              options={teamMembers}
              onValueChange={setSelectedMembers}
              placeholder="Select team members..."
              useSelectAll={true}
              maxDisplayedLabels={2}
            />
            <div className="text-sm space-y-2">
              <div>Selected ({selectedMembers.length}):</div>
              <div className="flex flex-wrap gap-1">
                {selectedMembers.map((id) => {
                  const member = teamMembers.find((m) => m.value === id);
                  return (
                    <Badge key={id} variant="secondary">
                      {member?.label}
                    </Badge>
                  );
                })}
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>MultiSelect - Languages</CardTitle>
            <CardDescription>
              Default select all enabled with custom init value
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <MultiSelect
              label="Languages"
              options={languages.map((lang) => ({ label: lang, value: lang.toLowerCase() }))}
              onValueChange={setSelectedLanguages}
              placeholder="Select languages..."
              defaultSelectAll={false}
              useSelectAll={true}
              maxDisplayedLabels={3}
              initValue={['typescript', 'javascript']}
            />
            <div className="text-sm space-y-2">
              <div>Selected ({selectedLanguages.length}):</div>
              <div className="flex flex-wrap gap-1">
                {selectedLanguages.map((lang) => (
                  <Badge key={lang} variant="outline">
                    {lang}
                  </Badge>
                ))}
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Usage Example Code */}
      <Card>
        <CardHeader>
          <CardTitle>✨ Key Features</CardTitle>
          <CardDescription>
            These components are now pure Controlled Components
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4 text-sm">
            <div className="space-y-2">
              <h3 className="font-semibold">✅ Benefits:</h3>
              <ul className="list-disc list-inside space-y-1 text-muted-foreground">
                <li>No internal API calls - fully controlled by parent</li>
                <li>Reusable across core, extensions, and shared packages</li>
                <li>No circular dependencies - can be moved to shared</li>
                <li>Clear separation: Presentation vs Container components</li>
                <li>Type-safe with proper TypeScript interfaces</li>
              </ul>
            </div>

            <div className="space-y-2">
              <h3 className="font-semibold">📦 Import:</h3>
              <pre className="bg-muted p-3 rounded-md overflow-x-auto">
{`import { SelectBox, MultiSelect } from '@pharos/shared/components';`}
              </pre>
            </div>

            <div className="space-y-2">
              <h3 className="font-semibold">💡 Usage Pattern:</h3>
              <pre className="bg-muted p-3 rounded-md overflow-x-auto text-xs">
{`// For dynamic data from API
const { query: { data } } = useList({
  resource: 'api/users',
  dataProviderName: 'default',
});

const options = data?.data?.map(item => ({
  label: item.name,
  value: item.id,
})) || [];

<SelectBox
  options={options}
  value={selected}
  onChange={setSelected}
/>`}
              </pre>
            </div>
          </div>
        </CardContent>
      </Card>
      </div>
    </div>
  );
};
