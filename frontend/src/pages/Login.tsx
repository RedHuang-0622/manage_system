import { useState } from 'react';
import { Form, Input, Button } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useAuth } from '../hooks/useAuth';
import { Navigate } from 'react-router-dom';
import { useAuthStore } from '../store/auth';

export default function Login() {
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const isLoggedIn = useAuthStore((s) => s.isLoggedIn());

  // Already logged in? Redirect to dashboard.
  // Use <Navigate> (declarative) instead of useNavigate() (imperative) to
  // avoid React's "cannot update BrowserRouter while rendering" warning.
  if (isLoggedIn) {
    return <Navigate to="/" replace />;
  }

  const onFinish = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      await login(values.username, values.password);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-container">
      <div className="login-card">
        <h1>实验室管理系统</h1>
        <Form name="login" size="large" onFinish={onFinish} autoComplete="off">
          <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input prefix={<UserOutlined />} placeholder="用户名" />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block>
              登录
            </Button>
          </Form.Item>
        </Form>
        <p style={{ textAlign: 'center', color: 'var(--color-text-secondary)', fontSize: 13 }}>
          默认管理员: admin / admin123
        </p>
      </div>
    </div>
  );
}
