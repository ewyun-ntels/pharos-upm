import {RefObject} from "react";
import {useOutsideclickRegistry} from "@hooks/use-outsideclick-registry";

let registeredRefArray: RefObject<HTMLDivElement | null>[] = [];

const useCloseSheet = () => {
  return useOutsideclickRegistry(registeredRefArray)
}

export {useCloseSheet}