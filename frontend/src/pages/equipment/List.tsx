import { useSearchParams } from 'react-router-dom';
import { Table, Button, Input, Select, Space, Card } from 'antd';
import { PlusOutlined, SearchOutlined } from '@ant-design/icons';
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { listEquipments } from '../../api/equipment';
import { usePagination } from '../../hooks/usePagination';
import { usePermission } from '../../hooks/usePermission';
import StatusBadge from '../../components/StatusBadge';
import type { Equipment } from '../../api/types';

export default function EquipList() {
  const navigate = useNavigate();
  const { isEquipManager } = usePermission();
  const pag = usePagination({ defaultPageSize: 12 });
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<Equipment[]>([]);
  const [searchParams] = useSearchParams();
  const [keyword, setKeyword] = useState(searchParams.get('keyword') || '');
  const [category, setCategory] = useState(searchParams.get('category') || '');
  const [onlyAvailable, setOnlyAvailable] = useState(searchParams.get('only_available') || '0');

  const fetchData = async (page: number, size: number) => {
    setLoading(true);
    try {
      const resp = await listEquipments({
        page,
        page_size: size,
        keyword: keyword || undefined,
        category: category || undefined,
        only_available: Number(onlyAvailable) as 0 | 1,
      });
      if (resp.code === 0 && resp.data) {
        setData(resp.data.list);
        pag.setTotal(resp.data.total);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchData(pag.page, pag.pageSize); }, [pag.page, pag.pageSize]);

  const handleSearch = () => {
    if (pag.page === 1) {
      fetchData(1, pag.pageSize);
    } else {
      pag.reset();
    }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '名称', dataIndex: 'name', ellipsis: true },
    { title: '型号', dataIndex: 'model', ellipsis: true },
    { title: '分类', dataIndex: 'category' },
    {
      title: '库存',
      key: 'stock',
      render: (_: unknown, r: Equipment) => (
        <span>
          <span style={{ color: '#52c41a', fontWeight: 600 }}>{r.available_stock}</span>
          {' / '}
          {r.total_stock}
        </span>
      ),
    },
    { title: '位置', dataIndex: 'location' },
    {
      title: '状态',
      dataIndex: 'status',
      render: (v: number) => <StatusBadge type="equipment" value={v} />,
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, r: Equipment) => (
        <Space>
          <Button type="link" size="small" onClick={() => navigate(`/equipments/${r.id}`)}>
            详情
          </Button>
          {isEquipManager && (
            <Button type="link" size="small" onClick={() => navigate(`/equipments/${r.id}/edit`)}>
              编辑
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <>
      <div className="page-header">
        <h2>设备大厅</h2>
      </div>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Space wrap>
          <Input
            placeholder="搜索设备名称"
            prefix={<SearchOutlined />}
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            onPressEnter={handleSearch}
            style={{ width: 200 }}
            allowClear
          />
          <Input
            placeholder="分类"
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            style={{ width: 150 }}
            allowClear
          />
          <Select
            value={onlyAvailable}
            onChange={(v) => { setOnlyAvailable(v); }}
            style={{ width: 130 }}
            options={[
              { value: '0', label: '全部状态' },
              { value: '1', label: '仅看有库存' },
            ]}
          />
          <Button type="primary" onClick={handleSearch} icon={<SearchOutlined />}>
            搜索
          </Button>
          {isEquipManager && (
            <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/equipments/new')}>
              设备入库
            </Button>
          )}
        </Space>
      </Card>
      <Table
        rowKey="id"
        columns={columns}
        dataSource={data}
        loading={loading}
        pagination={pag.paginationProps}
        scroll={{ x: 900 }}
      />
    </>
  );
}
