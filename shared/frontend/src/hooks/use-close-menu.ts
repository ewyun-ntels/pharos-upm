import {RefObject} from 'react';
import {useOutsideclickRegistry} from './use-outsideclick-registry';

let registeredRefArray: RefObject<HTMLDivElement | null>[] = [];

export const useCloseMenu = () => {
  return useOutsideclickRegistry(registeredRefArray);
};
