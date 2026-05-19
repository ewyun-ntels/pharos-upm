/**
 * Example Page Component
 */

import * as React from 'react';

export const ExamplePage: React.FC = () => {
  return (
    <div className="p-6">
      <h1 className="text-3xl font-bold mb-4">Example Extension Page</h1>
      <p className="text-gray-600 mb-4">
        This is an example page from the extension system.
      </p>
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <h2 className="text-xl font-semibold mb-2">Features:</h2>
        <ul className="list-disc list-inside space-y-1">
          <li>SITE_MODE-based conditional loading</li>
          <li>pnpm workspace integration</li>
          <li>Runtime registration system</li>
          <li>Hot module replacement support</li>
        </ul>
      </div>
    </div>
  );
};
