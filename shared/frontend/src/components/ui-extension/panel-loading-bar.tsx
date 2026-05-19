'use client';

export function PanelLoadingBar() {
  return (
    <>
      <style>{`
        @keyframes pharos-panel-loading {
          0%   { transform: translateX(-100%); }
          100% { transform: translateX(350%); }
        }
      `}</style>
      <div className="absolute top-0 left-0 right-0 h-[2px] overflow-hidden z-10 pointer-events-none">
        <div
          className="h-full w-[30%] bg-primary/70 rounded-full"
          style={{ animation: 'pharos-panel-loading 1.5s ease-in-out infinite' }}
        />
      </div>
    </>
  );
}
