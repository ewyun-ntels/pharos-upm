import React from 'react';
import {render, screen, fireEvent, waitFor} from '@testing-library/react';
import OptionsPanel from './options';
import {useDashboardStore, useDashboardData} from '@features/dashboard/hooks/use-dashboard-store';

// ─── 외부 의존성 Mock ─────────────────────────────────────────────────────────

const mockNavigate = jest.fn();
jest.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

jest.mock('@features/dashboard/hooks/use-dashboard-store', () => ({
  useDashboardStore: jest.fn(),
  useDashboardData: jest.fn(),
}));

// panelPluginRegistry.loadEditorConfig가 null을 반환하면 toPanelData 분기를 타지 않음
jest.mock('@/features/dashboard/panels/registry/PanelPluginRegistry', () => ({
  panelPluginRegistry: {
    getAllInfoAsSelectOptions: jest.fn(() => []),
    loadEditorConfig: jest.fn(() => Promise.resolve(null)),
  },
}));

jest.mock('@hooks/use-toast', () => ({
  useToast: () => ({toast: jest.fn()}),
}));

jest.mock('lucide-react', () => ({
  Save: () => <span>SaveIcon</span>,
  XCircle: () => <span>XCircleIcon</span>,
}));

// UI 컴포넌트 - 클릭/입력이 실제로 동작하도록 최소 래핑
jest.mock('@pharos/shared/components/ui/button', () => ({
  Button: ({children, onClick, disabled}: any) => (
    <button onClick={onClick} disabled={disabled}>
      {children}
    </button>
  ),
}));

jest.mock('@pharos/shared/components/ui/input', () => ({
  Input: ({value, onChange, id}: any) => (
    <input id={id} value={value || ''} onChange={onChange} />
  ),
}));

jest.mock('@pharos/shared/components/ui/textarea', () => ({
  Textarea: ({value, onChange, id}: any) => (
    <textarea id={id} value={value || ''} onChange={onChange} />
  ),
}));

jest.mock('@pharos/shared/components/ui/switch', () => ({
  Switch: ({checked, onCheckedChange}: any) => (
    <input
      type="checkbox"
      checked={checked}
      onChange={(e) => onCheckedChange(e.target.checked)}
    />
  ),
}));

jest.mock('@pharos/shared/components/ui-extension/select/box', () => ({
  SelectBox: ({value, onChange}: any) => (
    <select value={value} onChange={(e) => onChange(e.target.value)}>
      <option value="timeSeries">timeSeries</option>
    </select>
  ),
}));

// ─── 공통 픽스처 ──────────────────────────────────────────────────────────────

const DASHBOARD_ID = 'dashboard-1';
const PANEL_ID = 'panel-1';

const mockPanel = {
  id: PANEL_ID,
  title: 'Test Panel',
  renderType: 'timeSeries',
  description: 'desc',
  bgTransparent: false,
  options: {},
  layout: {type: 'card', x: 0, y: 0, w: 6, h: 4},
};

const mockUpdatePanel = jest.fn();
// saveDashboard가 store에는 존재하지만 options.tsx에서 구독/호출되어선 안 됨
const mockSaveDashboard = jest.fn();

function setupStoreMock(panelExists = true) {
  (useDashboardStore as unknown as jest.Mock).mockImplementation((selector: any) =>
    selector({
      updatePanel: mockUpdatePanel,
      saveDashboard: mockSaveDashboard,
      panelMap: panelExists ? {[PANEL_ID]: mockPanel} : {},
    }),
  );
  (useDashboardData as unknown as jest.Mock).mockReturnValue({permission: 'editor'});
}

function renderOptions(panelId = PANEL_ID) {
  return render(
    <OptionsPanel
      panelId={panelId}
      dashboardId={DASHBOARD_ID}
      queries={[]}
    />,
  );
}

// ─── 테스트 ───────────────────────────────────────────────────────────────────

describe('OptionsPanel', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    setupStoreMock();
  });

  describe('Save 버튼', () => {
    it('updatePanel을 호출하고 대시보드로 navigate해야 한다', async () => {
      renderOptions();

      fireEvent.click(screen.getByRole('button', {name: /save/i}));

      await waitFor(() => {
        expect(mockUpdatePanel).toHaveBeenCalledWith(
          PANEL_ID,
          expect.objectContaining({id: PANEL_ID}),
        );
        expect(mockNavigate).toHaveBeenCalledWith(`/dashboards/${DASHBOARD_ID}`);
      });
    });

    // 핵심: 이 테스트가 깨지면 누군가 패널 에디터에서 백엔드 저장을 다시 추가한 것
    it('saveDashboard를 호출하면 안 된다 — 백엔드 저장은 대시보드 Save에서만 처리', async () => {
      renderOptions();

      fireEvent.click(screen.getByRole('button', {name: /save/i}));

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalled(); // save 완료 확인
      });

      expect(mockSaveDashboard).not.toHaveBeenCalled();
    });
  });

  describe('Cancel 버튼', () => {
    it('store를 건드리지 않고 대시보드로 navigate해야 한다', () => {
      renderOptions();

      fireEvent.click(screen.getByRole('button', {name: /cancel/i}));

      expect(mockNavigate).toHaveBeenCalledWith(`/dashboards/${DASHBOARD_ID}`);
      expect(mockUpdatePanel).not.toHaveBeenCalled();
      expect(mockSaveDashboard).not.toHaveBeenCalled();
    });
  });

  describe('패널 없음', () => {
    it('panelData가 없으면 Panel not found 메시지를 표시해야 한다', () => {
      setupStoreMock(false);
      renderOptions();

      expect(screen.getByText(/Panel not found/i)).toBeInTheDocument();
    });

    it('panelData가 없으면 Save/Cancel 버튼이 렌더링되지 않아야 한다', () => {
      setupStoreMock(false);
      renderOptions();

      expect(screen.queryByRole('button', {name: /save/i})).not.toBeInTheDocument();
      expect(screen.queryByRole('button', {name: /cancel/i})).not.toBeInTheDocument();
    });
  });
});
