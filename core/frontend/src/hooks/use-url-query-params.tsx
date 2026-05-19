import {useNavigate, useSearchParams} from 'react-router-dom';

type QueryParams = Record<string, string | string[]>;

export const useUrlQueryParams = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  // 쿼리 파라미터를 객체 형태로 반환
  const getQueryParams = (): QueryParams => {
    const params: QueryParams = {};
    searchParams?.forEach((value, key) => {
      if (params[key] || params[key] === '') {
        // 동일한 키에 여러 값이 있는 경우 배열로 처리
        params[key] = Array.isArray(params[key])
          ? [...(params[key] as string[]), value]
          : [params[key] as string, value];
      } else {
        params[key] = value;
      }
    });

    return params;
  };

  // Update query parameters
  const setQueryParams = (newParams: QueryParams) => {
    const params = new URLSearchParams();

    Object.entries(newParams).forEach(([key, value]) => {
      if (Array.isArray(value)) {
        value.forEach((v) => params.append(key, v));
      } else {
        params.set(key, value);
      }
    });


    navigate(`${window.location.pathname.replace('/ui', '')}?${params.toString()}`);
  };

  return {getQueryParams, setQueryParams};
};
