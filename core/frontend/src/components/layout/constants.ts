
const defaultEditorSize = 500;

/** sideBar & Header 
 * layout 디자인은 3가지 타입임, 아래가 모두 false일때가 default이고, 아래를 하나씩 true로 변경하면 각각의 레이아웃에 해당하도록 변경됨 
 */
/* header */
const headerSize = 48
/* SNB */
const sidebarSize = 216 // Sidebar width
const iconMenuXSize = 8 //manuXPadding, minWidth 조정시 필수로 수동 조정필요
const menuXPadding = 12 //NavMenu main-menu.tsx의 메뉴 전체 컨테이너 px-[]값

/* control editor */
const controlHeight = 36;
const minHeight = controlHeight + headerSize; //하단 control editor의 위치값. (h-[100%] - minHeight인 위치 계산)
/* data-table size */
const thHeight = 8
const tdHeight = 9

export {defaultEditorSize, headerSize, sidebarSize, menuXPadding, iconMenuXSize, controlHeight, minHeight, thHeight, tdHeight};