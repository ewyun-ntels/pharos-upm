# pharos-frontend 설정 방식
- refine 및 tailwindcss, shadcn-ui 설치

```shell
✔ Downloaded remote source successfully.
✔ Choose a project template · refine-nextjs
✔ What would you like to name your project?: · n-component-template
✔ Choose your backend service to connect: · data-provider-custom-json-rest
✔ Do you want to use a UI Framework?: · no
✔ Do you want to add example pages?: · no
✔ Do you need any Authentication logic?: · none
✔ Choose a package manager: · npm
✔ Mind sharing your email? (We reach out to developers for free priority support, events, and SWAG kits. We never spam.) ·

Success! Created n-component-template at /Users/loykin/Project/n-component-template 🚀

npm install -D tailwindcss postcss autoprefixer

npx tailwindcss init -p

npx shadcn-ui@latest init
✔ Would you like to use TypeScript (recommended)? … no / yes
✔ Which style would you like to use? › New York
✔ Which color would you like to use as base color? › Slate
✔ Where is your global CSS file? … src/styles/globals.css
✔ Would you like to use CSS variables for colors? … no / yes
✔ Are you using a custom tailwind prefix eg. tw-? (Leave blank if not) …
✔ Where is your tailwind.config.js located? … tailwind.config.js
✔ Configure the import alias for components: … @/components
✔ Configure the import alias for utils: … @/lib/utils
✔ Are you using React Server Components? … no / yes
✔ Write configuration to components.json. Proceed? … yes
```

- tailwindcss css 변경
```
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
    :root {
        --background: 0 0% 100%;
        --foreground: 222.2 84% 4.9%;

        --card: 0 0% 100%;
        --card-foreground: 222.2 84% 4.9%;

        --popover: 0 0% 100%;
        --popover-foreground: 222.2 84% 4.9%;

        --primary: 222.2 47.4% 11.2%;
        --primary-foreground: 210 40% 98%;

        --secondary: 210 40% 96.1%;
        --secondary-foreground: 222.2 47.4% 11.2%;

        --muted: 210 40% 96.1%;
        --muted-foreground: 215.4 16.3% 46.9%;

        --accent: 210 40% 96.1%;
        --accent-foreground: 222.2 47.4% 11.2%;

        --destructive: 0 84.2% 60.2%;
        --destructive-foreground: 210 40% 98%;

        --border: 214.3 31.8% 91.4%;
        --input: 214.3 31.8% 91.4%;
        --ring: 222.2 84% 4.9%;

        --radius: 0.5rem;
    }

    .dark {
        --background: 222.2 84% 4.9%;
        --foreground: 210 40% 98%;

        --card: 222.2 84% 4.9%;
        --card-foreground: 210 40% 98%;

        --popover: 222.2 84% 4.9%;
        --popover-foreground: 210 40% 98%;

        --primary: 210 40% 98%;
        --primary-foreground: 222.2 47.4% 11.2%;

        --secondary: 217.2 32.6% 17.5%;
        --secondary-foreground: 210 40% 98%;

        --muted: 217.2 32.6% 17.5%;
        --muted-foreground: 215 20.2% 65.1%;

        --accent: 217.2 32.6% 17.5%;
        --accent-foreground: 210 40% 98%;

        --destructive: 0 62.8% 30.6%;
        --destructive-foreground: 210 40% 98%;

        --border: 217.2 32.6% 17.5%;
        --input: 217.2 32.6% 17.5%;
        --ring: 212.7 26.8% 83.9%;
    }
}

@layer base {
    * {
        @apply border-border;
    }

    body {
        @apply bg-background text-foreground;
    }

    .layout {
        @apply flex flex-col gap-1 md:flex-row;
    }

    .layout .content {
        @apply flex flex-grow flex-col;
    }

    .breadcrumb {
        @apply flex gap-6 border-b p-2;
        list-style-type: "/  ";
    }

    .breadcrumb a {
        @apply text-primary;
        text-decoration: none;
    }

    .menu {
        @apply mx-2 flex items-center justify-between border-gray-400 bg-slate-100 px-1 py-2 md:max-h-fit md:min-h-screen md:min-w-48 md:max-w-48 md:flex-col md:items-start md:justify-start md:px-0;
    }

    .menu a {
        @apply text-gray-950;
    }

    .menu .active {
        @apply font-bold;
    }

    .menu ul {
        @apply flex pl-4 md:flex-col md:justify-end;
    }
}
```
# i18n 적용 안내

프로젝트 내 다국어 지원(i18n) 적용 사항에 대해 설명합니다. 현재 다국어 지원은 좌측 메뉴와 Kubernetes > Monitoring 메뉴에 적용되어 있습니다.

## 다국어 관련 파일 경로

1. **엑셀 파일 경로**  
   다국어 관련 엑셀 파일은 다음 경로에 위치해 있습니다:  
   `NWDAF > 파일 > 03.개발 > NWDAF 다국어 관련 리스트`

2. **JSON 파일 경로**  
   각 언어별 JSON 파일은 아래 경로에 위치해 있습니다:
   - 영어: `/locales/en/translation.json`
   - 한국어: `/locales/ko/translation.json`

## i18n 적용 방법

페이지 내에서 다국어를 적용할 때는 다음과 같은 방식으로 사용합니다:

1. `@refinedev/core`의 `useTranslation` 훅을 import 합니다:
   ```javascript
   import { useTranslation } from "@refinedev/core";
   const { translate: t } = useTranslation();
   t("label.monitoring.refresh_interval")


