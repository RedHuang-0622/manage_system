import { Pagination as AntPagination } from 'antd';
import type { PaginationProps } from 'antd';

interface Props extends PaginationProps {}

/**
 * Reusable pagination wrapper around Ant Design Pagination.
 */
export default function Pagination(props: Props) {
  return (
    <AntPagination
      showSizeChanger
      showQuickJumper
      showTotal={(total) => `共 ${total} 条`}
      pageSizeOptions={['10', '20', '50', '100']}
      {...props}
    />
  );
}
