import { Table, Select, Space, Card, Empty, Alert } from 'antd';
import { useState } from 'react';
import { listMyRecords, cancelBorrow } from '../../api/borrows';
import { usePagination } from '../../hooks/usePagination';
import { useRequest } from '../../hooks/useRequest';
import StatusBadge from '../../components/StatusBadge';
import type { BorrowRecord } from '../../api/types';
import { Button, message, Popconfirm } from 'antd';
import { AxiosError } from 'axios';
import { ErrCode } from '../../api/types';

export default function MyRecords() {
  const pag = usePagination();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<BorrowRecord[]>([]);
  const [status, setStatus] = useState<string>('');
  const [error, setError] = useState<string | null>(null);

  const fetchData = async () => {
    setLoading(true);
    setError(null);
    try {
      const resp = await listMyRecords({ page: pag.page, page_size: pag.pageSize, status: status || undefined });
      if (resp.code === 0 && resp.data) {
        setData(resp.data.list);
        pag.setTotal(resp.data.total);
      }
    } catch (err: unknown) {
      const axiosErr = err as AxiosError<{ code: number; msg: string }>;
      if (axiosErr.response?.status === 401 && axiosErr.response?.data?.code === ErrCode.ErrTokenInvalid) {
        // 拦截器已处理 refresh，这里不应到达；若到达说明 refresh 也失败了
        return; // 拦截器会 redirect to /login
      }
      setError(axiosErr.message || '加载失败，请检查网络连接');
    } finally {
      setLoading(false);
    }
  };

  useRequest(() => { fetchData(); }, [pag.page, pag.pageSize, status]);

  const handleCancel = async (id: number) => {
    try {
      const resp = await cancelBorrow(id);
      if (resp.code === 0) {
        message.success('已取消申请');
        fetchData();
      } else {
        message.error(resp.msg || '取消失败');
      }
    } catch { message.error('操作失败'); }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    {
      title: '设备',
      key: 'equipment',
      render: (_: unknown, r: BorrowRecord) => r.equipment ? `${r.equipment.name} (${r.equipment.model})` : `设备#${r.equipment_id}`,
    },
    { title: '数量', dataIndex: 'quantity', width: 60 },
    { title: '申请备注', dataIndex: 'apply_note', ellipsis: true },
    { title: '审批备注', dataIndex: 'approve_note', ellipsis: true },
    {
      title: '状态',
      dataIndex: 'status',
      render: (v: string) => <StatusBadge type="borrow" value={v} />,
    },
    { title: '申请时间', dataIndex: 'apply_at', width: 170 },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_: unknown, r: BorrowRecord) =>
        r.status === '申请中' ? (
          <Popconfirm title="确定取消申请？" onConfirm={() => handleCancel(r.id)}>
            <Button type="link" size="small" danger>取消</Button>
          </Popconfirm>
        ) : null,
    },
  ];

  return (
    <>
      <div className="page-header"><h2>我的借阅</h2></div>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Space>
          <Select
            value={status}
            onChange={(v) => { setStatus(v); }}
            style={{ width: 150 }}
            allowClear
            placeholder="筛选状态"
            options={[
              { value: '申请中', label: '申请中' },
              { value: '已借出', label: '已借出' },
              { value: '已归还', label: '已归还' },
              { value: '被拒绝', label: '被拒绝' },
            ]}
          />
        </Space>
      </Card>
      {error ? (
        <Alert type="error" message="加载失败" description={error} showIcon style={{ marginBottom: 16 }} />
      ) : null}
      <Table
        rowKey="id"
        columns={columns}
        dataSource={data}
        loading={loading}
        pagination={pag.paginationProps}
        scroll={{ x: 900 }}
        locale={{ emptyText: <Empty description="暂无借阅记录" /> }}
      />
    </>
  );
}
