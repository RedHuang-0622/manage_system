import { Table, Button, Input, Space, message, Popconfirm, Empty, Alert } from 'antd';
import { useEffect, useState } from 'react';
import { listPendingRecords, approveBorrow } from '../../api/borrows';
import { usePagination } from '../../hooks/usePagination';
import StatusBadge from '../../components/StatusBadge';
import type { BorrowRecord } from '../../api/types';
import { AxiosError } from 'axios';
import { ErrCode } from '../../api/types';

export default function PendingList() {
  const pag = usePagination();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<BorrowRecord[]>([]);
  const [note, setNote] = useState('');
  const [error, setError] = useState<string | null>(null);

  const fetchData = async () => {
    setLoading(true);
    setError(null);
    try {
      const resp = await listPendingRecords({ page: pag.page, page_size: pag.pageSize });
      if (resp.code === 0 && resp.data) {
        setData(resp.data.list);
        pag.setTotal(resp.data.total);
      }
    } catch (err: unknown) {
      const axiosErr = err as AxiosError<{ code: number; msg: string }>;
      if (axiosErr.response?.status === 401 && axiosErr.response?.data?.code === ErrCode.ErrTokenInvalid) {
        return;
      }
      setError(axiosErr.message || '加载失败');
    } finally { setLoading(false); }
  };

  useEffect(() => { fetchData(); }, [pag.page, pag.pageSize]);

  const handleApprove = async (id: number, approve: boolean, approveNote: string) => {
    try {
      const resp = await approveBorrow(id, { approve, approve_note: approveNote });
      if (resp.code === 0) {
        message.success(approve ? '审批通过' : '已拒绝');
        fetchData();
      } else {
        message.error(resp.msg || '操作失败');
      }
    } catch { message.error('操作失败'); }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    {
      title: '申请人',
      key: 'user',
      render: (_: unknown, r: BorrowRecord) => r.user ? `${r.user.real_name} (${r.user.username})` : `用户#${r.user_id}`,
    },
    {
      title: '设备',
      key: 'equipment',
      render: (_: unknown, r: BorrowRecord) => r.equipment ? `${r.equipment.name} (${r.equipment.model})` : `设备#${r.equipment_id}`,
    },
    { title: '数量', dataIndex: 'quantity', width: 60 },
    { title: '申请备注', dataIndex: 'apply_note', ellipsis: true },
    {
      title: '状态',
      dataIndex: 'status',
      render: (v: string) => <StatusBadge type="borrow" value={v} />,
    },
    { title: '申请时间', dataIndex: 'apply_at', width: 170 },
    {
      title: '操作',
      key: 'action',
      width: 180,
      render: (_: unknown, r: BorrowRecord) => (
        <Space>
          <Popconfirm
            title="审批通过"
            description={<Input placeholder="审批备注（可选）" onChange={(e) => setNote(e.target.value)} />}
            onConfirm={() => handleApprove(r.id, true, note)}
            okText="确认通过"
          >
            <Button type="link" size="small" style={{ color: '#52c41a' }}>通过</Button>
          </Popconfirm>
          <Popconfirm
            title="拒绝申请"
            description={<Input placeholder="拒绝原因" onChange={(e) => setNote(e.target.value)} />}
            onConfirm={() => handleApprove(r.id, false, note)}
            okText="确认拒绝"
            okButtonProps={{ danger: true }}
          >
            <Button type="link" size="small" danger>拒绝</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <div className="page-header"><h2>待审批列表</h2></div>
      {error ? <Alert type="error" message="加载失败" description={error} showIcon style={{ marginBottom: 16 }} /> : null}
      <Table rowKey="id" columns={columns} dataSource={data} loading={loading} pagination={pag.paginationProps} scroll={{ x: 1000 }}
        locale={{ emptyText: <Empty description="暂无待审批工单" /> }} />
    </>
  );
}
