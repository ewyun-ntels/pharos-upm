import { RefObject } from "react";

// outside click시 이벤트 감지 로직을 관리하는 hook

const useOutsideclickRegistry = (registeredRefArray: RefObject<HTMLDivElement | null>[]) => {
  const outSideClickDisable = (ref: RefObject<HTMLDivElement | null>) => {
    registeredRefArray.push(ref);
  };

  // 성능상 문제가 있어보이는 로직이긴 하지만 일단 영향이 갈정도로 길진 않을것이기때문에 놔둠
  // 봐서 링크드리스트로 변경이 필요할지도
  const outSideClickEnable = (ref: RefObject<HTMLDivElement | null>) => {
    const index = registeredRefArray.indexOf(ref);
    if (index > -1) {
      registeredRefArray.splice(index, 1);
    }
  }

  const getRegistryRefArray = () => {
    return registeredRefArray;
  }

  return {
    outSideClickDisable,
    outSideClickEnable,
    getRegistryRefArray
  };
};

export { useOutsideclickRegistry }
