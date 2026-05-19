import {
  ChevronLeftIcon,
  ChevronRightIcon,
  DoubleArrowLeftIcon,
  DoubleArrowRightIcon,
} from '@radix-ui/react-icons';

import {Button} from '@pharos/shared/components/ui';

interface DataPaginationProps {
  currentPage: number;
  totalPage: number;
  totalCount: number;
  onPageChange: (page: number) => void;
  isLoading?: boolean;
}

export function DataPagination({
  currentPage,
  totalPage,
  totalCount,
  onPageChange,
  isLoading = false,
}: DataPaginationProps) {
  const isNoDataPage = totalCount === 0 && totalPage === 0;

  const goToFirstPage = () => {
    onPageChange(1);
  };

  const goToPreviousPage = () => {
    if (currentPage > 1) {
      onPageChange(currentPage - 1);
    }
  };

  const goToNextPage = () => {
    if (currentPage < totalPage) {
      onPageChange(currentPage + 1);
    }
  };

  const goToLastPage = () => {
    onPageChange(totalPage);
  };

  const getCurrentPage = () => {
    if (isNoDataPage) {
      return 1;
    }

    return currentPage;
  };

  return (
    <div className="pagination flex items-center justify-start ">
      <div className="flex items-center space-x-2">
        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            className="hidden h-8 w-8 p-0 lg:flex"
            onClick={goToFirstPage}
            disabled={isLoading || isNoDataPage || currentPage === 1}
          >
            <span className="sr-only">Go to first page</span>
            <DoubleArrowLeftIcon className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            className="h-8 w-8 p-0"
            onClick={goToPreviousPage}
            disabled={isLoading || isNoDataPage || currentPage <= 1}
          >
            <span className="sr-only">Go to previous page</span>
            <ChevronLeftIcon className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            className="h-8 w-8 p-0"
            onClick={goToNextPage}
            disabled={isLoading || currentPage >= totalPage}
          >
            <span className="sr-only">Go to next page</span>
            <ChevronRightIcon className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            className="hidden h-8 w-8 p-0 lg:flex"
            onClick={goToLastPage}
            disabled={isLoading || currentPage >= totalPage}
          >
            <span className="sr-only">Go to last page</span>
            <DoubleArrowRightIcon className="h-4 w-4" />
          </Button>
        </div>
        <div className="flex items-end justify-center text-xs text-muted-foreground ">
          {`${getCurrentPage().toLocaleString()} / ${totalPage !== 0 ? totalPage.toLocaleString() : 1}`}
        </div>
      </div>
    </div>
  );
}
