import { useState } from 'react';

interface UsePaginationOptions {
  defaultPage?: number;
  defaultPageSize?: number;
}

export function usePagination(options: UsePaginationOptions = {}) {
  const { defaultPage = 1, defaultPageSize = 10 } = options;
  const [page, setPage] = useState(defaultPage);
  const [pageSize, setPageSize] = useState(defaultPageSize);
  const [total, setTotal] = useState(0);

  const onChange = (newPage: number, newPageSize: number) => {
    setPage(newPage);
    setPageSize(newPageSize);
  };

  const reset = () => {
    setPage(defaultPage);
    setPageSize(defaultPageSize);
    setTotal(0);
  };

  return {
    page,
    pageSize,
    total,
    setTotal,
    onChange,
    reset,
    // Ready-to-use Ant Design Table pagination props
    paginationProps: {
      current: page,
      pageSize,
      total,
      showSizeChanger: true,
      showTotal: (t: number) => `共 ${t} 条`,
      onChange,
    },
  };
}
