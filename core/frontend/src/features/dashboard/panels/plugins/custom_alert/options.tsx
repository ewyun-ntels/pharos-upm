import {Accordion} from '@pharos/shared/components/ui-extension';
import type {PanelEditorOptionsProps} from '@pharos/core/panel-registry';
import {CustomAlertPanelOptions} from './types';
import {Input} from '@pharos/shared/components/ui';
import {useOptionsChange} from '../common/useOptionsChange';

export const Options: React.FC<PanelEditorOptionsProps<CustomAlertPanelOptions>> = ({
  options = {},
  onOptionsChange,
  info,
}) => {
  const handleChange = useOptionsChange(options, onOptionsChange);

  return (
    <div>
      <Accordion
        id={info?.id || 'custom_alert'}
        items={[
          {
            id: 'alertOptions',
            title: 'Alert options',
            content: (
              <div className="space-y-4">
                {/* Custom message */}
                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">
                    Chart data not exist message
                  </label>
                  <Input
                    value={options.chartDataNotExistMessage || ''}
                    onChange={(e) => handleChange(d => { d.chartDataNotExistMessage = e.target.value; })}
                    placeholder="Enter custom message"
                  />
                </div>
              </div>
            ),
          },
        ]}
        defaultOpenItems={'all'}
        type="multiple"
      />
    </div>
  );
};
