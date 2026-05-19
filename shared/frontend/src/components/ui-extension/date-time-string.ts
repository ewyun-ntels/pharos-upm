
// ko: 2024. 10. 7. 오후 2:25  en: 10/7/2024, 2:25 PM 형식으로 locale별 자동으로 표현식 변경됨
const defaultDateFiledOptions: Intl.DateTimeFormatOptions = {
    year: "numeric",
    month: "numeric",
    day: "numeric",
    hour: "numeric",
    minute: "numeric"
};

// DateFieldStringProps
interface DateFieldStringProps {
    //locale 지정하지 않을 경우 기본 locale(useLocal) 사용
    locale?: string;
    // date 지정하지 않을 경우 현재 시간 사용
    date?: Date;
    // options 지정하지 않을 경우 defaultDateFiledOptions 사용
    options?: Intl.DateTimeFormatOptions;
}

function DateFieldString({
                             locale,
                             date = new Date(),
                             options = defaultDateFiledOptions
                         }: DateFieldStringProps): string {
    // navigator에서 직접 가져옴 useLocale() 사용시 Rules of Hooks 에러 발생 가능
    const effectiveLocale = locale ??  navigator.language;

    return new Intl.DateTimeFormat(effectiveLocale, options).format(date);

    // 한국어 형식인 경우만 날짜 뒤의 마지막 `.`을 빈칸으로 바꾸는 표현식 주석 처리함
    // const formattedDate = new Intl.DateTimeFormat(effectiveLocale, options).format(date);

    // return effectiveLocale.startsWith('ko')
    //     ? formattedDate.replace(/(\d{1,2})\.(\s오[후|전])/, '$1 $2')
    //     : formattedDate;
}

export { DateFieldString, defaultDateFiledOptions };
export type { DateFieldStringProps };
