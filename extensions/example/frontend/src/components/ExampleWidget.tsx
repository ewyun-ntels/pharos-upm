/**
 * Example Widget Component
 */

import * as React from 'react';

interface ExampleWidgetProps {
  title?: string;
}

export const ExampleWidget: React.FC<ExampleWidgetProps> = ({ 
  title = 'Example Widget' 
}) => {
  return (
    <div className="border rounded-lg p-4 bg-white shadow-sm">
      <h3 className="text-lg font-semibold mb-2">{title}</h3>
      <p className="text-sm text-gray-600">
        This widget is loaded from the example extension.
      </p>
      <div className="mt-4 flex items-center justify-between">
        <span className="text-xs text-gray-500">Extension: example</span>
        <span className="px-2 py-1 bg-green-100 text-green-800 text-xs rounded">
          Active
        </span>
      </div>
    </div>
  );
};
