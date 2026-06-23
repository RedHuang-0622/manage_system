import { Table, Select, Input, Space, Card, Button, message, Empty, Alert } from 'antd';
import { useState } from 'react';
import { useRequest } from '../../hooks/useRequest';
import { listAllRecords, returnBorrow } from '../../api/borrows';
import { usePagination } from '../../hooks/usePagination';
import StatusBadge from '../../components/StatusBadge';
import type { BorrowRecord } from '../../api/types';
import { AxiosError } from 'axios';
import { ErrCode } from '../../api/types';

export default function AllRecords() {
  const pag = usePagination();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<BorrowRecord[]>([]);
  const [status, setStatus] = useState<string>('');
  const [userId, setUserId] = useState('');
  const [error, setError] = useState<string | null>(null);

  const fetchData = async () => {
    setLoading(true);
    setError(null);
    try {
      const resp = await listAllRecords({
        page: pag.page,
        page_size: pag.pageSize,
        status: status || undefined,
        user_id: userId ? Number(userId) : undefined,
      });
      if (resp.code === 0 && resp.data) {
        setData(resp.data.list);
        pag.setTotal(resp.data.total);
      }
    } catch (err: unknown) {
      const axiosErr = err as AxiosError<{ code: number; msg: string }>;
      if (axiosErr.response?.status === 401 && axiosErr.response?.data?.code === ErrCode.ErrTokenInvalid) {
        return; // interceptor handles redirect to /login
      }
      setError(axiosErr.message || '加载失败，请检查网络连接');
    } finally { setLoading(false); }
  };

  useRequest(() => { fetchData(); }, [pag.page, pag.pageSize, status, userId]);

  const handleReturn = async (id: number) => {
    try {
      const resp = await returnBorrow(id);
      if (resp.code === 0) {
        message.success('归还成功');
        fetchData();
      } else {
        message.error(resp.msg || '归还失败');
      }
    } catch { message.error('操作失败'); }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    {
      title: '借用人',
      key: 'user',
      render: (_: unknown, r: BorrowRecord) => r.user ? `${r.user.real_name} (${r.user.username})` : `用户#${r.user_id}`,
    },
    {
      title: '设备',
      key: 'equipment',
      render: (_: unknown, r: BorrowRecord) => r.equipment ? `${r.equipment.name} (${r.equipment.model})` : `设备#${r.equipment_id}`,
    },
    { title: '数量', dataIndex: 'quantity', width: 60 },
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
        r.status === '已借出' ? (
          <Button type="link" size="small" onClick={() => handleReturn(r.id)}>归还</Button>
        ) : null,
    },
  ];

  return (
    <>
      <div className="page-header"><h2>全部借阅记录</h2></div>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Space wrap>
          <Select
            value={status}
            onChange={(v) => setStatus(v)}
            style={{ width: 120 }}
            allowClear
            placeholder="状态筛选"
            options={[
              { value: '申请中', label: '申请中' },
              { value: '已借出', label: '已借出' },
              { value: '已归还', label: '已归还' },
              { value: '被拒绝', label: '被拒绝' },
            ]}
          />
          <Input placeholder="借用人ID" value={userId} onChange={(e) => setUserId(e.target.value)} style={{ width: 120 }} allowClear />
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
