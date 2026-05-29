import { registerExtension } from '@pharos/core/extension-registry';
import type { Extension } from '@pharos/core/extension-registry';
import HomeUPMPage from './routes/home-upm';

const metadata: Extension = {
  name: 'home-upm',
  version: '1.0.0',
  displayName: 'Kubernetes Home',
  description: 'UPM Kubernetes cluster home page',

  pages: [
    {
      path: '/extensions/home-upm',
      component: HomeUPMPage,
      title: 'Home',
      requireAuth: true,
    },
  ],
};

registerExtension(metadata);

export {metadata};
export {default as HomeUPMPage} from './routes/home-upm';
