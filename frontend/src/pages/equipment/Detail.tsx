import { useParams, useNavigate } from 'react-router-dom';
import { Descriptions, Button, Spin, Card, Space, Tag } from 'antd';
import { ArrowLeftOutlined, EditOutlined } from '@ant-design/icons';
import { useEffect, useState } from 'react';
import { getEquipment } from '../../api/equipment';
import { usePermission } from '../../hooks/usePermission';
import StatusBadge from '../../components/StatusBadge';
import type { Equipment } from '../../api/types';

export default function EquipDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { isEquipManager } = usePermission();
  const [loading, setLoading] = useState(true);
  const [equip, setEquip] = useState<Equipment | null>(null);

  useEffect(() => {
    (async () => {
      if (!id) return;
      setLoading(true);
      try {
        const resp = await getEquipment(Number(id));
        if (resp.code === 0) setEquip(resp.data);
      } finally {
        setLoading(false);
      }
    })();
  }, [id]);

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '40px auto' }} />;
  if (!equip) return <div>设备不存在</div>;

  return (
    <>
      <div className="page-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2>{equip.name}</h2>
        <Space>
          {isEquipManager && (
            <Button icon={<EditOutlined />} onClick={() => navigate(`/equipments/${id}/edit`)}>
              编辑
            </Button>
          )}
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/equipments')}>
            返回列表
          </Button>
        </Space>
      </div>
      <Card>
        <Descriptions bordered column={{ xs: 1, sm: 2 }}>
          <Descriptions.Item label="ID">{equip.id}</Descriptions.Item>
          <Descriptions.Item label="名称">{equip.name}</Descriptions.Item>
          <Descriptions.Item label="型号">{equip.model || '-'}</Descriptions.Item>
          <Descriptions.Item label="分类">{equip.category || '-'}</Descriptions.Item>
          <Descriptions.Item label="总库存">{equip.total_stock}</Descriptions.Item>
          <Descriptions.Item label="可用库存">
            <Tag color={equip.available_stock > 0 ? 'green' : 'red'}>{equip.available_stock}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="位置">{equip.location || '-'}</Descriptions.Item>
          <Descriptions.Item label="状态"><StatusBadge type="equipment" value={equip.status} /></Descriptions.Item>
          <Descriptions.Item label="创建时间">{equip.created_at}</Descriptions.Item>
          <Descriptions.Item label="更新时间">{equip.updated_at}</Descriptions.Item>
          <Descriptions.Item label="描述" span={2}>{equip.description || '-'}</Descriptions.Item>
        </Descriptions>
      </Card>
    </>
  );
}
