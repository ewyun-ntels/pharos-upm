import {create} from 'zustand';
import {subscribeWithSelector} from 'zustand/middleware';
import {AttachAddon} from '@xterm/addon-attach';

// WebSocket 상태 타입 정의
interface ChannelState {
  websocket?: WebSocket;
  attachaddon: AttachAddon;
  messages: any[]; // 채널에서 수신한 메시지
}

type TerminalState = {
  channels: Record<string, ChannelState>; // 채널별 WebSocket 상태
  logs: string[]; // 로그 데이터
  commandInput?: string;
  initializeChannel: (url: string, channel: string) => void;
  handleChannelMessage: (channel: string, data: any) => void;
  removeChannel: (channel: string) => void;
  //sendCommandToChannel: (channel: string, command: string, payload?: Record<string, any>) => void;
  closeAllChannels: () => void;
};

// WebSocket 생성 함수
const createChannelWebSocket = (
  url: string,
  channel: string,
  onMessage: (channel: string, data: any) => void,
  onError: (channel: string, error: Event) => void,
  onOpen?: () => void,
  onClose?: (channel: string) => void,
): WebSocket => {
  const ws = new WebSocket(url);

  ws.onopen = () => {
    console.log(`WebSocket connected for channel: ${channel}`);
    //ws.send(JSON.stringify({ action: "subscribe", channel }));
    onOpen?.();
  };

  ws.onmessage = (event: MessageEvent) => {
    try {
      // const message = event.data;
      // if (message !== "# ") {
      //     onMessage(channel, message);
      // }
      const message = event.data;
      onMessage(channel, message);
    } catch (error) {
      console.error(`Invalid message on channel ${channel}:`, event.data);
    }
  };

  ws.onerror = (error) => {
    console.error(`WebSocket error on channel ${channel}:`, error);
    onError(channel, error);
  };

  ws.onclose = () => {
    console.log(`WebSocket disconnected for channel: ${channel}`);
    onClose?.(channel);
  };

  return ws;
};

// Zustand Store 정의
const useTerminalStore = create(
  subscribeWithSelector<TerminalState>((set, get) => ({
    channels: {}, // 채널별 상태
    logs: [], // 로그 데이터
    commandInput: '', // Channel 에 보낸 명령어
    // 채널 WebSocket 초기화
    initializeChannel: (url: string, channel: string) => {
      const existingChannel = get().channels[channel];
      if (existingChannel) return; // 이미 채널 WebSocket이 존재하면 실행 안 함

      const ws = createChannelWebSocket(
        url,
        channel,
        (channel, data) => get().handleChannelMessage(channel, data),
        (channel, error) => console.error(`Error on channel ${channel}:`, error),
        () => console.log(`Channel ${channel} connected`),
        (channel) => get().removeChannel(channel),
      );

      const attachAddon = new AttachAddon(ws);

      set((state) => ({
        channels: {
          ...state.channels,
          [channel]: {websocket: ws, attachaddon: attachAddon, messages: []},
        },
      }));
    },

    // 채널 메시지 처리
    handleChannelMessage: (channel: string, data: string) => {
      set((state) => ({
        channels: {
          ...state.channels,
          [channel]: {
            ...state.channels[channel],
            messages: [...state.channels[channel].messages, data],
          },
        },
      }));
    },

    handleChannelClose: (channel: string) => {
      set((state) => ({
        channels: {
          ...state.channels,
          [channel]: {
            ...state.channels[channel],
            messages: [...state.channels[channel].messages, 'WebSocket disconnected for channel'],
          },
        },
      }));
    },

    // 채널 제거
    removeChannel: (channel: string) => {
      const attachAddonChannels = get().channels;
      if (attachAddonChannels[channel]?.websocket) {
        attachAddonChannels[channel].websocket.close();
      }
      delete attachAddonChannels[channel];
    },

    // 모든 채널 WebSocket 닫기
    closeAllChannels: () => {
      const attachAddonChannels = get().channels;
      Object.keys(attachAddonChannels).forEach((channel) => {
        if (attachAddonChannels[channel]?.websocket) {
          attachAddonChannels[channel].websocket.close();
        }
      });
      set({channels: {}});
    },
  })),
);

export type {TerminalState};
export default useTerminalStore;
