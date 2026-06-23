import { useNavigate } from 'react-router-dom';
import { Form, Input, InputNumber, Button, Card, message } from 'antd';
import { useState } from 'react';
import { applyBorrow } from '../../api/borrows';
import type { ApplyBorrowReq } from '../../api/types';

export default function BorrowApply() {
  const navigate = useNavigate();
  const [form] = Form.useForm<ApplyBorrowReq>();
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: ApplyBorrowReq) => {
    setLoading(true);
    try {
      const resp = await applyBorrow(values);
      if (resp.code === 0) {
        message.success('申请已提交，等待审批');
        navigate('/borrows/my');
      } else {
        message.error(resp.msg || '申请失败');
      }
      } catch {
        message.error('网络异常，请检查网络连接');
      } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <div className="page-header"><h2>发起借阅申请</h2></div>
      <Card style={{ maxWidth: 500 }}>
        <Form form={form} layout="vertical" onFinish={onFinish}>
          <Form.Item name="equipment_id" label="设备ID" rules={[{ required: true, message: '请输入设备ID' }]}>
            <InputNumber min={1} style={{ width: '100%' }} placeholder="可在设备大厅查看设备ID" />
          </Form.Item>
          <Form.Item name="quantity" label="借阅数量" rules={[{ required: true, message: '请输入借阅数量' }]}>
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="apply_note" label="申请备注">
            <Input.TextArea maxLength={256} rows={3} placeholder="说明借阅用途" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading}>提交申请</Button>
            <Button style={{ marginLeft: 8 }} onClick={() => navigate('/equipments')}>去设备大厅看看</Button>
          </Form.Item>
        </Form>
      </Card>
    </>
  );
}
