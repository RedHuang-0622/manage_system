import { useParams, useNavigate } from 'react-router-dom';
import { Form, Input, Select, Button, Card, Spin, message } from 'antd';
import { useEffect, useState } from 'react';
import { getUser, updateUser } from '../../api/users';
import { listRoles } from '../../api/auth';
import type { UpdateUserReq, RoleInfo } from '../../api/types';

export default function UserEdit() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form] = Form.useForm<UpdateUserReq>();
  const [loading, setLoading] = useState(false);
  const [fetching, setFetching] = useState(true);
  const [roles, setRoles] = useState<RoleInfo[]>([]);

  useEffect(() => {
    (async () => {
      if (!id) return;
      setFetching(true);
      try {
        const [userResp, rolesResp] = await Promise.all([
          getUser(Number(id)),
          listRoles(),
        ]);
        if (userResp.code === 0) {
          form.setFieldsValue(userResp.data);
        } else {
          message.error(userResp.msg || '获取用户信息失败');
        }
        if (rolesResp.code === 0) {
          setRoles(rolesResp.data);
        } else {
          message.error(rolesResp.msg || '获取角色列表失败');
        }
      } catch {
        message.error('网络异常，请检查网络连接');
      } finally {
        setFetching(false);
      }
    })();
  }, [id, form]);

  const onFinish = async (values: UpdateUserReq) => {
    if (!id) return;
    setLoading(true);
    try {
      const resp = await updateUser(Number(id), values);
      if (resp.code === 0) {
        message.success('更新成功');
        navigate('/users');
      } else {
        message.error(resp.msg || '更新失败');
      }
    } catch {
      message.error('网络异常，请检查网络连接');
    } finally {
      setLoading(false);
    }
  };

  if (fetching) return <Spin size="large" style={{ display: 'block', margin: '40px auto' }} />;

  return (
    <>
      <div className="page-header"><h2>编辑用户</h2></div>
      <Card style={{ maxWidth: 600 }}>
        <Form form={form} layout="vertical" onFinish={onFinish}>
          <Form.Item name="real_name" label="真实姓名" rules={[{ min: 2, max: 32 }]}><Input maxLength={32} /></Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ type: 'email' }]}><Input maxLength={64} /></Form.Item>
          <Form.Item name="phone" label="手机号"><Input maxLength={16} /></Form.Item>
          <Form.Item name="role_id" label="角色">
            <Select options={roles.map(r => ({ value: r.id, label: `${r.role_name} — ${r.description}` }))} allowClear />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select options={[{ value: 1, label: '启用' }, { value: 0, label: '禁用' }]} allowClear />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading}>保存</Button>
            <Button style={{ marginLeft: 8 }} onClick={() => navigate('/users')}>取消</Button>
          </Form.Item>
        </Form>
      </Card>
    </>
  );
}
