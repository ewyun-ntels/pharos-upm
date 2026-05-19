import { uiConfigProvider, UI_CONFIG_RESOURCES } from '@providers/ui-config-provider';
import {useEffect, useState} from 'react';


export function useConfig() {
    const [config, setConfig] = useState<{[key: string]: any}>({"style-primary": "oklch(0.66 0.1197 156.75)"});

    useEffect(() => {
        const defaultStyle = () => {
            uiConfigProvider.getList({
                resource: UI_CONFIG_RESOURCES.CONFIG,
                pagination: undefined,
            }).then((response: any) => {
                const data = response.data;
                if (data) {
                    const configInfo: {[key: string]: any} = data;
                    setConfig(configInfo);
                } else {
                    setConfig({"style-primary": "oklch(0.66 0.1197 156.75)"});
                }
            }).catch((error: any) => {
                console.error('Error fetching config:', error);
                setConfig({"style-primary": "oklch(0.66 0.1197 156.75)"});
            });
        }

        defaultStyle();
    }, []);

    return { config }
}
