import React from "react";
// import {Editor} from "@components/editor";
import {TerminalView} from "@components/terminal/Terminal";
// import {LogView} from "@components/logView";
import { TabEditor, TabEditorData, useTabStore } from "@components/controlBar/controlTabStore";
import { ResourceLoader } from "@components/editor/resourceLoader";

type TabContentProps = {
    tabId: string
    height: number;
    onResize: (ySize: number, enable: boolean) => void;
};

const TabContent: React.FC<TabContentProps> = ({tabId, onResize}) => {
    const tabStore = useTabStore();
    const tabData = tabStore.getAny(tabId)

    // const onCloseHandler = (id: string) => {
    //   tabStore.remove(id)
    //   onResize(0, false)
    // }

    switch (tabData?.type) {
        case 'editor':
        case 'modifyEditor':
          const editData: TabEditor = tabData

          const onSaveData = (id: string, saveData: TabEditorData) => {
            let data = tabStore.tabData.get(id)
            const tabType = data?.type as 'editor' | 'modifyEditor'
            if (data && (data.type === 'editor' || data.type === 'modifyEditor')) {
              tabStore.setEditor(id, data.type, data.typeId, data.label, undefined, undefined, saveData);
            } else {
              console.error(`save data error: ${tabType}`)
            }
          }

          if (editData.data.showProps) {
            // - ResourceLoader는 dataProvider를 통해 Data를 받기 위해 사용되며 Data를 정상적으로 수신할경우 onSaveData를 호출후 종료함
            // - onSaveData는 수신한 데이터를 tabStore에 저장후 re-rendering 발생
            // - ResourceLoader에서 onSaveData를 호출하면서 data.showProps를 제거하기 때문에 하단의 Editor가 호출됨
            // - 차후 ResourceLoader안에 Editor를 넣어서 사용할 수 있지만, re-rander를 활용하는 것이 낮다고 판단되어 Editor를 분리하여 사용함
            return <ResourceLoader
              id={tabId}
              showProps={editData.data.showProps}
              dataLoader={onSaveData}
            />
          } else if (editData.data.data !== undefined) {
            return <div>Editor Chart sample</div>
            // return <Editor key={tabId} // unmount하기 위해 사용
            //                tabType={editData.type}
            //                language={"yaml"}
            //                height={height}
            //                width={"100%"}
            //                id={tabId}
            //                onCloseHandler={onCloseHandler}
            //                data={editData.data.data}
            //                line={editData?.line}
            //                column={editData?.column}
            //                 onSaveData={onSaveData}
            //                onSaveLine={(id, line, column) => {
            //                  let data = tabStore.tabData.get(id)
            //                  if (data && (data.type === 'editor' || data.type === 'modifyEditor')) {
            //                    // 이전 값을 무조건 덮어씌움
            //                    tabStore.setEditor(id, data.type, data.typeId, data.label, line, column, data.data)
            //                  } else {
            //                    console.error(`save line error: ${data?.type}`)
            //                  }
            //                }}
            // />
          } else {
            return <div>{`Editor Data is Empty(${editData.data})`}</div>
          }
        case 'terminal':
            const websocketUrl = `ws://${window.location.host}/kubernetes/exec/${tabData.namespace}/${tabData.pod}?container=${tabData.selectContainer}`;

            return <TerminalView id={tabId}
                            data={{
                                namespace: tabData.namespace,
                                pod: tabData.pod,
                                selectContainer: tabData.selectContainer,
                            }}
                            websocketUrl={websocketUrl}
                            selectedContainer={(id, container) => tabStore.setTerminal(id,
                                tabData.typeId,
                                tabData.label,
                                tabData.namespace,
                                tabData.pod,
                                container,
                            )}
                            tabRemove={(tabId: string) => tabStore.remove(tabId)}
                            tabDataSize={tabStore.tabData.size}
                            onResize={onResize}
                    />
        case 'logView':
            return <div>LogView sample</div>
            // return <LogView id={tabId} height={height} width={"100%"}
            //                 selectedContainer={(id, container) => tabStore.setLogView(id,
            //                     tabData.typeId,
            //                     tabData.label,
            //                     tabData.namespace,
            //                     tabData.pod,
            //                     container,
            //                     tabData.containers
            //                 )}
            //                 data={{
            //                     namespace: tabData.namespace,
            //                     pod: tabData.pod,
            //                     selectContainer: tabData.selectContainer,
            //                     containers: tabData.containers,
            //                 }}/>
        default:
            return <div>Unknown Tab Id = {tabId}</div>
    }
}

export {TabContent}