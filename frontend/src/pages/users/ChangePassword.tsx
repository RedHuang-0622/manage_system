import { useParams, useNavigate } from 'react-router-dom';
import { Form, Input, Button, Card, message } from 'antd';
import { useState } from 'react';
import { changePassword } from '../../api/users';
import type { ChangePasswordReq } from '../../api/types';

export default function ChangePassword() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form] = Form.useForm<ChangePasswordReq>();
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: ChangePasswordReq) => {
    if (!id) return;
    setLoading(true);
    try {
      const resp = await changePassword(Number(id), values);
      if (resp.code === 0) {
        message.success('密码修改成功');
        navigate(-1);
      } else {
        message.error(resp.msg || '修改失败');
      }
    } catch {
      message.error('网络异常，请检查网络连接');
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <div className="page-header"><h2>修改密码</h2></div>
      <Card style={{ maxWidth: 500 }}>
        <Form form={form} layout="vertical" onFinish={onFinish}>
          <Form.Item name="old_password" label="原密码" rules={[{ required: true, min: 6, max: 64 }]}>
            <Input.Password maxLength={64} />
          </Form.Item>
          <Form.Item name="new_password" label="新密码" rules={[{ required: true, min: 6, max: 64 }]}>
            <Input.Password maxLength={64} />
          </Form.Item>
          <Form.Item
            name="confirm"
            label="确认新密码"
            dependencies={['new_password']}
            rules={[
              { required: true, message: '请确认新密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('new_password') === value) return Promise.resolve();
                  return Promise.reject(new Error('两次输入的密码不一致'));
                },
              }),
            ]}
          >
            <Input.Password maxLength={64} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading}>确认修改</Button>
            <Button style={{ marginLeft: 8 }} onClick={() => navigate(-1)}>取消</Button>
          </Form.Item>
        </Form>
      </Card>
    </>
  );
}
