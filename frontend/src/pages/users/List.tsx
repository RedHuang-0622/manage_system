import { Table, Button, Input, Select, Space, Card } from 'antd';
import { PlusOutlined, SearchOutlined } from '@ant-design/icons';
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { listUsers, disableUser } from '../../api/users';
import { usePagination } from '../../hooks/usePagination';
import { usePermission } from '../../hooks/usePermission';
import StatusBadge from '../../components/StatusBadge';
import type { User } from '../../api/types';
import { message, Popconfirm } from 'antd';

export default function UserList() {
  const navigate = useNavigate();
  const { isSuperAdmin } = usePermission();
  const pag = usePagination();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<User[]>([]);
  const [keyword, setKeyword] = useState('');
  const [status, setStatus] = useState<string>('-1');

  const fetchData = async () => {
    setLoading(true);
    try {
      const resp = await listUsers({
        page: pag.page,
        page_size: pag.pageSize,
        keyword: keyword || undefined,
        status: Number(status),
      });
      if (resp.code === 0 && resp.data) {
        setData(resp.data.list);
        pag.setTotal(resp.data.total);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchData(); }, [pag.page, pag.pageSize]);

  const handleSearch = () => { pag.reset(); fetchData(); };

  const handleDisable = async (id: number) => {
    try {
      const resp = await disableUser(id);
      if (resp.code === 0) {
        message.success('用户已禁用');
        fetchData();
      } else {
        message.error(resp.msg || '禁用失败');
      }
    } catch { message.error('操作失败'); }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '用户名', dataIndex: 'username' },
    { title: '姓名', dataIndex: 'real_name' },
    { title: '邮箱', dataIndex: 'email', ellipsis: true },
    { title: '手机', dataIndex: 'phone' },
    {
      title: '角色',
      dataIndex: ['role', 'role_name'],
      render: (v: string) => {
        const labels: Record<string, string> = { super_admin: '超级管理员', lab_admin: '实验室负责人', member: '普通成员' };
        return labels[v] || v;
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      render: (v: number) => <StatusBadge type="user" value={v} />,
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, r: User) => (
        <Space>
          <Button type="link" size="small" onClick={() => navigate(`/users/${r.id}/edit`)}>编辑</Button>
          {isSuperAdmin && r.status !== 0 && (
            <Popconfirm title="确定禁用该用户？" onConfirm={() => handleDisable(r.id)}>
              <Button type="link" size="small" danger>禁用</Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <>
      <div className="page-header"><h2>用户管理</h2></div>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Space wrap>
          <Input placeholder="搜索用户名/姓名" prefix={<SearchOutlined />} value={keyword}
            onChange={(e) => setKeyword(e.target.value)} onPressEnter={handleSearch} style={{ width: 200 }} allowClear />
          <Select value={status} onChange={setStatus} style={{ width: 120 }}
            options={[{ value: '-1', label: '全部' }, { value: '1', label: '启用' }, { value: '0', label: '禁用' }]} />
          <Button type="primary" onClick={handleSearch} icon={<SearchOutlined />}>搜索</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/users/new')}>创建用户</Button>
        </Space>
      </Card>
      <Table rowKey="id" columns={columns} dataSource={data} loading={loading} pagination={pag.paginationProps} scroll={{ x: 800 }} />
    </>
  );
}
