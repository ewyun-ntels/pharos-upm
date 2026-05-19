
// import { useEffect, useState } from 'react';
// import {useList, useForm} from '@/lib/data-provider';
// import {useToast} from '@hooks/use-toast';
// import {UI_CONFIG_PROVIDER_NAME, UI_CONFIG_RESOURCES} from '@providers/ui-config-provider';
//
// export default function SetThemePage() {
//   const [configData, setConfigData] = useState<any>(null);
//   const [primaryColor, setPrimaryColor] = useState<string>('');
//
//   useEffect(() => {
//     const root = document.documentElement;
//     const cssPrimary = getComputedStyle(root).getPropertyValue('--primary').trim();
//
//     // 초기값 설정
//     setPrimaryColor(cssPrimary);
//   }, []);
//
//   // 값이 설정되면 CSS 변수 업데이트
//   useEffect(() => {
//     if (primaryColor) {
//       document.documentElement.style.setProperty('--primary', primaryColor);
//     }
//   }, [primaryColor]);
//
//   const {query: {data}} = useList<{[key: string]: any}>({
//     resource: UI_CONFIG_RESOURCES.CONFIG,
//     dataProviderName: UI_CONFIG_PROVIDER_NAME,
//   });
//
//   // 설정 값 가져오기
//   useEffect(() => {
//     if (!data || !data.data) return;
//     const config = data.data as Record<string, any>;
//     setConfigData(config);
//     // 초기값 설정
//     setPrimaryColor(config["style-primary"] || 'oklch(0.66 0.1197 156.75)');
//   }, [data]);
//
//   const {onFinish} = useForm({
//     action: 'edit',
//     resource: UI_CONFIG_RESOURCES.CONFIG,
//     dataProviderName: UI_CONFIG_PROVIDER_NAME,
//     redirect: false,
//   });
//
//   const {toast} = useToast();
//   const onSubmit = () => {
//     onFinish({
//       ...configData,
//       "style-primary": primaryColor,
//     })
//       .then(() => {
//         toast({description: 'Successfully saved.'});
//       })
//       .catch((error) => {
//         // toast({variant: 'destructive', description: 'error'});
//         console.error('Error:', error);
//       });
//   };
//
//   return (
//     <div className="space-y-6 p-8 max-w-lg">
//       <h1 className="text-2xl font-bold">Primary 색상 설정</h1>
//
//       <label className="block space-y-2">
//         <span className="text-sm font-medium text-muted-foreground">색상 값 입력</span>
//         <input
//           type="text"
//           value={primaryColor}
//           onChange={(e) => setPrimaryColor(e.target.value)}
//           placeholder="{primaryColor}"
//           className="w-full border px-3 py-2 rounded"
//         />
//       </label>
//
//       <label className="block space-y-2">
//       <p className="text-sm text-muted-foreground">컬러 선택</p>
//         <input
//           type="color"
//           value={primaryColor}
//           onChange={(e) => {
//             const hex = e.target.value;
//             document.documentElement.style.setProperty('--primary', hex);
//             setPrimaryColor(hex);
//           }}
//           className="w-16 h-10 p-0 border-none bg-transparent"
//         />
//       </label>
//
//       <div className="space-y-2">
//         <p className="text-sm text-muted-foreground">미리보기</p>
//         <div className="w-full h-12 rounded bg-primary text-primary-foreground flex items-center justify-center font-medium">
//           bg-primary
//         </div>
//         <button className="bg-primary text-primary-foreground px-4 py-2 rounded">버튼 예시</button>
//       </div>
//       <div className='space-y-2'>
//         <p className="text-sm text-muted-foreground">설정 정보 저장</p>
//         <button onClick={()=> onSubmit()} className="bg-primary text-primary-foreground px-4 py-2 rounded cursor-pointer">저장</button>
//       </div>
//     </div>
//   );
// }
