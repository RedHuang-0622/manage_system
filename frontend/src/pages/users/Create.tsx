import { useNavigate } from 'react-router-dom';
import { Form, Input, Select, Button, Card, message } from 'antd';
import { useEffect, useState } from 'react';
import { createUser } from '../../api/users';
import { listRoles } from '../../api/auth';
import type { CreateUserReq, RoleInfo } from '../../api/types';

export default function UserCreate() {
  const navigate = useNavigate();
  const [form] = Form.useForm<CreateUserReq>();
  const [loading, setLoading] = useState(false);
  const [roles, setRoles] = useState<RoleInfo[]>([]);

  useEffect(() => {
    (async () => {
      try {
        const resp = await listRoles();
        if (resp.code === 0) setRoles(resp.data);
      } catch { /* ignore */ }
    })();
  }, []);

  const onFinish = async (values: CreateUserReq) => {
    setLoading(true);
    try {
      const resp = await createUser(values);
      if (resp.code === 0) {
        message.success('用户创建成功');
        navigate('/users');
      } else {
        message.error(resp.msg || '创建失败');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <div className="page-header"><h2>创建用户</h2></div>
      <Card style={{ maxWidth: 600 }}>
        <Form form={form} layout="vertical" onFinish={onFinish}>
          <Form.Item name="username" label="用户名" rules={[{ required: true, min: 3, max: 32 }]}>
            <Input maxLength={32} />
          </Form.Item>
          <Form.Item name="real_name" label="真实姓名" rules={[{ required: true, min: 2, max: 32 }]}>
            <Input maxLength={32} />
          </Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, min: 6, max: 64 }]}>
            <Input.Password maxLength={64} />
          </Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ type: 'email', message: '邮箱格式不正确' }]}>
            <Input maxLength={64} />
          </Form.Item>
          <Form.Item name="phone" label="手机号"><Input maxLength={16} /></Form.Item>
          <Form.Item name="role_id" label="角色" rules={[{ required: true, message: '请选择角色' }]}>
            <Select options={roles.map(r => ({ value: r.id, label: `${r.role_name} — ${r.description}` }))} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading}>创建</Button>
            <Button style={{ marginLeft: 8 }} onClick={() => navigate('/users')}>取消</Button>
          </Form.Item>
        </Form>
      </Card>
    </>
  );
}
