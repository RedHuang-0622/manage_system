import { useParams, useNavigate } from 'react-router-dom';
import { Form, Input, InputNumber, Select, Button, Card, Spin, message } from 'antd';
import { useEffect, useState } from 'react';
import { getEquipment, updateEquipment, disableEquipment } from '../../api/equipment';
import { usePermission } from '../../hooks/usePermission';
import type { UpdateEquipReq } from '../../api/types';

export default function EquipEdit() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { isSuperAdmin } = usePermission();
  const [form] = Form.useForm<UpdateEquipReq>();
  const [loading, setLoading] = useState(false);
  const [fetching, setFetching] = useState(true);

  useEffect(() => {
    (async () => {
      if (!id) return;
      setFetching(true);
      try {
        const resp = await getEquipment(Number(id));
        if (resp.code === 0) {
          form.setFieldsValue(resp.data);
        }
      } finally {
        setFetching(false);
      }
    })();
  }, [id, form]);

  const onFinish = async (values: UpdateEquipReq) => {
    if (!id) return;
    setLoading(true);
    try {
      const resp = await updateEquipment(Number(id), values);
      if (resp.code === 0) {
        message.success('更新成功');
        navigate(`/equipments/${id}`);
      } else {
        message.error(resp.msg || '更新失败');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleDisable = async () => {
    if (!id) return;
    setLoading(true);
    try {
      const resp = await disableEquipment(Number(id));
      if (resp.code === 0) {
        message.success('设备已下架');
        navigate('/equipments');
      } else {
        message.error(resp.msg || '下架失败');
      }
    } finally {
      setLoading(false);
    }
  };

  if (fetching) return <Spin size="large" style={{ display: 'block', margin: '40px auto' }} />;

  return (
    <>
      <div className="page-header"><h2>编辑设备</h2></div>
      <Card style={{ maxWidth: 600 }}>
        <Form form={form} layout="vertical" onFinish={onFinish}>
          <Form.Item name="name" label="设备名称" rules={[{ required: true }]}><Input maxLength={128} /></Form.Item>
          <Form.Item name="model" label="型号"><Input maxLength={64} /></Form.Item>
          <Form.Item name="category" label="分类"><Input maxLength={32} /></Form.Item>
          <Form.Item name="total_stock" label="总库存"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="location" label="存放位置"><Input maxLength={64} /></Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea maxLength={1024} rows={3} /></Form.Item>
          <Form.Item name="status" label="状态">
            <Select options={[{ value: 1, label: '上架' }, { value: 0, label: '下架' }]} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading}>保存</Button>
            <Button style={{ marginLeft: 8 }} onClick={() => navigate(`/equipments/${id}`)}>取消</Button>
            {isSuperAdmin && (
              <Button danger style={{ marginLeft: 8 }} onClick={handleDisable}>下架设备</Button>
            )}
          </Form.Item>
        </Form>
      </Card>
    </>
  );
}
