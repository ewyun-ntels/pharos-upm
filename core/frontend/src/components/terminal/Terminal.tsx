import React, {useEffect, useRef } from 'react';
import useTerminalStore from '@components/terminal/terminalStore';
import {Terminal} from '@xterm/xterm';
import {FitAddon} from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';
import {useTheme} from '@providers/theme-provider';

type TerminalType = {
  namespace: string;
  pod: string;
  selectContainer: string;
};

interface TerminalProps {
  id: string;
  data: TerminalType;
  websocketUrl: string;
  selectedContainer: (tabId: string, container: string) => void;
  tabRemove: (tabId: string) => void;
  tabDataSize: number;
  onResize: (ySize: number, enable: boolean) => void;
}

const TerminalView: React.FC<TerminalProps> = ({
  id,
  data,
  websocketUrl,
  tabRemove,
  tabDataSize,
  onResize,
}) => {
  const terminalRef = useRef<HTMLDivElement | null>(null);
  const terminalInstance = useRef<Terminal | null>(null);
  const fitAddon = new FitAddon();

  //console.log("TerminalView channel ::[", websocketUrl, "]::[", id, "]==>", data);
  const initializeChannel = useTerminalStore((state) => state.initializeChannel); // init Channel
  //const sendCommandToChannel = useTerminalStore((state) => state.sendCommandToChannel); // command
  const removeChannel = useTerminalStore((state) => state.removeChannel); // remove Channel

  let commandLine = '';
  const channelName = `${id}-${data.pod}-${data.namespace}-${data.selectContainer}`;

  const {theme} = useTheme();

  useEffect(() => {
    if (!terminalRef.current) return;

    terminalInstance.current = new Terminal({
      theme:
        theme === 'dark'
          ? {
              background: '#171717',
              foreground: '#ffffff',
              cursor: '#f0f0f0',
              selectionBackground: '#444444',
              selectionForeground: '#ffffff',
            }
          : {
              background: '#ffffff',
              foreground: '#18181b',
              cursor: '#000000',
              selectionBackground: '#d0d0d0',
              selectionForeground: '#000000',
            },
      fontSize: 12,
      fontFamily: "Menlo, Monaco, Consolas, 'Courier New', monospace",
      cursorBlink: true,
    });
    terminalInstance.current.loadAddon(fitAddon);
    terminalInstance.current.open(terminalRef.current);

    // Attach the terminal to the DOM
    if (terminalRef.current) terminalInstance.current.open(terminalRef.current);

    const resizeObserver = new ResizeObserver(() => {
      fitAddon.fit(); // div 크기가 변경되면 터미널 크기 조정
    });

    resizeObserver.observe(terminalRef.current);
    //fitAddon.fit();

    // Set up WebSocket connection
    initializeChannel(websocketUrl, channelName);

    const attachAddon = useTerminalStore.getState().channels[channelName].attachaddon;

    if (attachAddon) terminalInstance.current.loadAddon(attachAddon);

    const messages = useTerminalStore.getState().channels[channelName].messages;
    if (messages.length === 0) {
      terminalInstance.current.writeln('Welcome to the WebSocket Terminal!');
      terminalInstance.current?.focus();
    } else {
      //console.log("메세지 :: ==>", messages);
      terminalInstance.current.writeln('Welcome to the WebSocket Terminal!');
      messages.forEach((message: string, index: number) => {
        if (messages.length - 1 === index) terminalInstance.current?.write(message);
        else terminalInstance.current?.write(message);
      });
      terminalInstance.current?.focus();
    }

    terminalInstance.current.onKey(({key, domEvent}) => {
      if (!useTerminalStore.getState().channels[channelName]) return;
      const printable = !domEvent.altKey && !domEvent.ctrlKey && !domEvent.metaKey;
      if (key === '\r') {
        // Enter key
        if (commandLine === 'exit') {
          removeChannel(channelName);

          terminalInstance.current?.dispose();
          tabRemove(id); // tab 삭제
          if (tabDataSize - 1 === 0) {
            onResize(0, false); //tabResize
          }
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
        commandLine = '';
      } else if (domEvent.key === 'Backspace') {
        if (commandLine === '') return;
        // Backspace key
        commandLine = commandLine.substring(0, commandLine.length - 1);
      } else if (printable) {
        commandLine += key;
      }
    });

    // 사용하지 않음.
    // const unsubscribe = useTerminalStore.subscribe(
    //     (state: TerminalState) => state.channels[channelName].messages,
    //     (newMessage: any) => {
    //         console.log("New message ::=>", newMessage);
    //     }
    // );

    return () => {
      //unsubscribe();
      terminalInstance.current?.dispose();
    };
  }, [theme, channelName, websocketUrl]);

  return (
    <div
      style={{display: 'flex', flexDirection: 'column', height: '100%'}}
      className="px-6 py-2 bg-background"
    >
      <div id="terminal" ref={terminalRef} style={{flexGrow: 1, height: '100%'}} />
    </div>
  );
};

export {TerminalView};
