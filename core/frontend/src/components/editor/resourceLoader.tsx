import {useShow, UseShowProps} from '@/lib/data-provider';
import {TabEditorData} from '@components/controlBar/controlTabStore';
import {useEffect} from 'react';
import * as yaml from 'js-yaml';

interface ResourceLoaderProps {
  // tabId 값으로 데이터를 저장할때 사용
  id: string;
  // useShow를 사용하기 위한 props
  showProps: UseShowProps;
  // loader에서 받은 data를 저장하기 위한 함수
  dataLoader: (id: string, saveData: TabEditorData) => void;
}

// ResourceLoader에서 받은 Data를 Editor에 뿌려주기 위해 onSaveData를 호출 후 종료
//  re-rander를 발생하기 위한 컴포넌트로 사용하기 때문에 재작성시 주의 필요
const ResourceLoader = ({id, showProps, dataLoader}: ResourceLoaderProps) => {
  const {query} = useShow(showProps);

  useEffect(() => {
    const rowData = query.data?.data;
    if (rowData) {
      dataLoader(id, {data: yaml.dump(rowData)});
    }
  }, [dataLoader, id, query.data?.data]);

  if (query.isLoading) {
    return <div>loading...</div>;
  }

  if (query.isError) {
    return <div>error...</div>;
  }

  return;
};

export {ResourceLoader};
