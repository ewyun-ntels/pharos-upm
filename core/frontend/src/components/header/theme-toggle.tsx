import {Moon, Sun} from '@pharos/shared/components';
import {useTheme} from '@providers/theme-provider';
import {IconButton} from '@pharos/shared/components/ui-extension';

export function ThemeToggle() {
  const {setTheme, theme} = useTheme();

  return (
    <IconButton
      variant="ghost"
      size="icon"
      className="rounded-sm"
      icon={
        <>
          <Sun className="h-4 w-4 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
          <Moon className="absolute top-2 h-4 w-4 rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
        </>
      }
      onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
    >
      Theme {/*TODO translate {t('label.common.theme')} */}
    </IconButton>
  );
}
