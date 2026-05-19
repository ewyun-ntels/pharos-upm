import {create} from "zustand";
import {persist} from "zustand/middleware";
import { UseShowProps } from '@/lib/data-provider';

// TODO TAB관련 모듈을 any 형으로 지정하여 플러그인형상으로 리팩토링 필요
// tab의 타입을 정의
type TabType = 'editor' | 'modifyEditor' | 'terminal' | 'logView'

interface TabBase {
    type: TabType
    typeId: number //tab id (각 타입별 마지막 번호를 id로 사용 하며 tab name의 마지막 번호를 확인하기 위해 사용)
    label: string // tab name
}

// data가 존재하지 않을경우 showProps를 사용하여 editor를 생성
type TabEditorData = {
    // editor에 표시할 data
    data?: string
    // editor를 생성하기 위한 showProps로 data가 없을경우 사용
    showProps?: UseShowProps
}

interface TabEditor extends TabBase {
    type: 'editor' | 'modifyEditor'
    line: number
    column: number
    data: TabEditorData
}

interface TabTerminal extends TabBase {
    type: 'terminal'
    namespace: string
    pod: string
    selectContainer: string
}

interface TabLog extends TabBase {
    type: 'logView'
    namespace: string
    pod: string
    selectContainer: string
    containers: string[]
}

type TabElement = TabEditor | TabTerminal | TabLog

const getNextTabId = (id: number) : string=> (id + 1).toString();

const getLabel = (type: TabType, id: number) => {
    switch (type) {
        case 'editor':
            return `Editor (${id})`
        case 'modifyEditor':
            return `Modify Editor (${id})`
        case 'terminal':
            return `Terminal (${id})`
        case 'logView':
            return `Log (${id})`
        default:
            return `Unknown (${id})`
    }
}

function getLastTypeId(type: TabType, tabsData: Map<string, TabElement>): number {
    let lastTypeId = 0;
    Array.from(tabsData.keys()).forEach((value) => {
        const data = tabsData.get(value)

        if (data?.type === type) {
            const typeNum = data.typeId;
            if (lastTypeId < typeNum) {
                lastTypeId = typeNum;
            }
        }
    });
    return lastTypeId;
}

function getNextTypeId(type: TabType, tabsData: Map<string, TabElement>): number {
    return getLastTypeId(type, tabsData) + 1;
}

// tab의 가장 마지막 번호 return
function getLastId(tabsData: Map<string, TabElement>): { tabType: string, lastId: number } {
    let lastId = 0;
    let lastTypeId = 0;
    let lastTabType = "";

    if (tabsData.size === 0) {
        return {tabType: "", lastId: 0}
    }

    Array.from(tabsData.keys()).forEach((value) => {
        const typeNum = tabsData.get(value)?.typeId ?? 0;
        if (lastTypeId < typeNum) {
            lastTypeId = typeNum;
        }

        const num = parseInt(value, 10);
        if (lastId < num) {
            lastId = num
        }

    });

    return {tabType: lastTabType, lastId: lastId,}
}

type TabStore = {
    tabData: Map<string, TabElement>
    lastSelectId: string | null
    setEditor: (id: string, type: 'editor' | 'modifyEditor', typeId: number, label: string, line: number | undefined, column: number | undefined, data: TabEditorData) => void
    getEditor: (id: string) => TabEditor | undefined;
    setTerminal: (id: string, typeId: number, label: string, namespace: string, pod: string, selectContainer: string) => void
    getTerminal: (id: string) => TabTerminal | undefined
    setLogView: (id: string, typeId: number, label: string, namespace: string, pod: string, selectContainer: string, containers: string[]) => void
    getLogView: (id: string) => TabLog | undefined
    getAny: (id: string) => TabElement | undefined
    remove: (id: string) => void
    setLastSelectId: (id: string) => void
}

// tab 내부의 component 데이터를 관리하기 위해 사용
const useTabStore = create<TabStore>()(
    persist(
        (set, get) => ({
            tabData: new Map<string, TabElement>(),
            lastSelectId: null,
            setEditor: (id: string, type: 'editor' | 'modifyEditor', typeId: number, label: string, line: number | undefined, column: number | undefined, data: TabEditorData) => set((state) => {
                    state.tabData.set(id, {
                        typeId: typeId,
                        label: label,
                        line: line ? line : 0,
                        column: column ? column : 0,
                        data: data,
                        type: type
                    })

                return {}
            }),
            getEditor: (id: string) => {
                const saveData = get().tabData.get(id);
                if (saveData?.type === 'editor' || saveData?.type === 'modifyEditor') {
                    return saveData as TabEditor;
                }
                return undefined
            },
            setTerminal: (id: string, typeId: number, label: string, namespace: string, pod: string, selectContainer: string) => set((state) => {
                // const saveData = state.tabData.get(id)
               
                state.tabData.set(id, {
                    typeId: typeId,
                    label: label,
                    namespace: namespace,
                    pod: pod,
                    selectContainer: selectContainer,
                    type: "terminal"
                })

                return {}
            }),
            getTerminal: (id: string) => {
                const saveData = get().tabData.get(id);
                if (saveData?.type === 'terminal') {
                    return saveData as TabTerminal;
                }
                return undefined
            },
            setLogView: (id: string, typeId: number, label: string, namespace: string, pod: string, selectContainer: string, containers: string[]) => set((state) => {
                // const saveData = state.tabData.get(id)
                // 중복 저장
                state.tabData.set(id, {
                    typeId: typeId,
                    label: label,
                    namespace: namespace,
                    pod: pod,
                    selectContainer: selectContainer,
                    containers: containers,
                    type: "logView"
                })

                return {}
            }),
            getLogView: (id: string) => {
                const saveData = get().tabData.get(id);
                if (saveData?.type === 'logView') {
                    return saveData as TabLog;
                }
                return undefined
            },
            getAny: (id: string) => {
                return get().tabData.get(id);
            },
            // tab 정보 삭제
            remove: (id: string) => set((state) => {
                state.tabData.delete(id)
                if (state.lastSelectId === id) {
                    const {lastId} = getLastId(state.tabData)
                    lastId === 0 ? state.lastSelectId = null :
                        // state.lastSelectId = tabType + "-" + lastId
                        state.lastSelectId = lastId.toString()
                }
                return {}
            }),
            // mount시 load해야할 tab id 지정
            setLastSelectId: (id: string) => set((state) => {
                state.lastSelectId = id
                return {}
            }),
        }), {
            name: "n-component-template-editor-tab-content",
            storage: {
                getItem: (name) => {
                    const str = localStorage.getItem(name);
                    if (str === null) return {state: {}};
                    return {
                        state: {
                            ...JSON.parse(str).state,
                            tabData: new Map(JSON.parse(str).state.tabData),
                        },
                    }
                },
                setItem: (name, newValue) => {
                    const str = JSON.stringify({
                        state: {
                            ...newValue.state,
                            tabData: Array.from(newValue.state.tabData.entries()),
                        },
                    })
                    localStorage.setItem(name, str)
                },
                removeItem: (name) => localStorage.removeItem(name),
            },
        })
)

export {useTabStore, getLastId, getNextTabId, getLastTypeId, getNextTypeId, getLabel};
export type {TabType, TabEditor, TabLog, TabStore, TabEditorData}