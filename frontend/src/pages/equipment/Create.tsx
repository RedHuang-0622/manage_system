import { useNavigate } from 'react-router-dom';
import { Form, Input, InputNumber, Button, Card, message } from 'antd';
import { useState } from 'react';
import { createEquipment } from '../../api/equipment';
import type { CreateEquipReq } from '../../api/types';

export default function EquipCreate() {
  const navigate = useNavigate();
  const [form] = Form.useForm<CreateEquipReq>();
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: CreateEquipReq) => {
    setLoading(true);
    try {
      const resp = await createEquipment(values);
      if (resp.code === 0) {
        message.success('设备入库成功');
        navigate('/equipments');
      } else {
        message.error(resp.msg || '创建失败');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <div className="page-header"><h2>设备入库</h2></div>
      <Card style={{ maxWidth: 600 }}>
        <Form form={form} layout="vertical" onFinish={onFinish} initialValues={{ status: 1 }}>
          <Form.Item name="name" label="设备名称" rules={[{ required: true, message: '请输入设备名称' }]}>
            <Input maxLength={128} />
          </Form.Item>
          <Form.Item name="model" label="型号"><Input maxLength={64} /></Form.Item>
          <Form.Item name="category" label="分类"><Input maxLength={32} placeholder="如: 服务器/网络设备/测量仪器" /></Form.Item>
          <Form.Item name="total_stock" label="总库存" rules={[{ required: true, message: '请输入库存数量' }]}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="location" label="存放位置"><Input maxLength={64} placeholder="如: A301实验室" /></Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea maxLength={1024} rows={3} /></Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading}>确认入库</Button>
            <Button style={{ marginLeft: 8 }} onClick={() => navigate('/equipments')}>取消</Button>
          </Form.Item>
        </Form>
      </Card>
    </>
  );
}
