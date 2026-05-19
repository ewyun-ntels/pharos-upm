import {create} from "zustand";
import { TabEditorData, TabType } from "./controlTabStore";

/*
 * 외부에서 ControlTab을 생성하는 이벤트를 관리하는 store
 * Sheet 같은 외부 Component에서 Tab 생성이 필요할경우 사용하며 queue 형태로 관리
 */

interface CreateBase {
    type: TabType;
    label?: string;
}

interface CreateEditor extends CreateBase {
    type: 'editor'| 'modifyEditor';
    data: TabEditorData
}

interface CreateTerminal extends CreateBase {
    type: 'terminal';
    namespace: string;
    pod: string;
    selectContainer: string;
}

interface CreateLogView extends CreateBase {
    type: 'logView';
    namespace: string;
    pod: string;
    selectContainer: string;
    containers: string[];
}

type CreateEvent = CreateEditor | CreateTerminal | CreateLogView;

type CreateEventStore = {
    eventData: CreateEvent[];
    setEvent: (event: CreateEvent) => void;
}

const useCreateEventStore = create<CreateEventStore>(set => ({
    eventData: [],
    setEvent: (event: CreateEvent) => set(state => ({
        eventData: [...state.eventData, event]
    })),
}));

export {useCreateEventStore}
export type {CreateEvent, CreateEditor, CreateTerminal, CreateLogView}