import { Button, Dropdown } from 'antd';
import { UserOutlined, LogoutOutlined, KeyOutlined } from '@ant-design/icons';
import { useAuth } from '../../hooks/useAuth';
import { useAuthStore } from '../../store/auth';
import { useNavigate } from 'react-router-dom';

export default function TopBar() {
  const { logout } = useAuth();
  const user = useAuthStore((s) => s.user);
  const navigate = useNavigate();

  const roleLabels: Record<string, string> = {
    super_admin: '超级管理员',
    lab_admin: '实验室负责人',
    member: '普通成员',
  };

  return (
    <div
      style={{
        height: 56,
        background: '#fff',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'flex-end',
        padding: '0 24px',
        borderBottom: '1px solid #f0f0f0',
        gap: 8,
      }}
    >
      <span style={{ color: 'var(--color-text-secondary)', fontSize: 13 }}>
        {roleLabels[user?.role_name || ''] || ''}
      </span>
      <Dropdown
        menu={{
          items: [
            {
              key: 'password',
              icon: <KeyOutlined />,
              label: '修改密码',
              onClick: () => navigate(`/users/${user?.user_id}/password`),
            },
            { type: 'divider' },
            {
              key: 'logout',
              icon: <LogoutOutlined />,
              label: '退出登录',
              onClick: logout,
            },
          ],
        }}
      >
        <Button type="text" icon={<UserOutlined />}>
          {user?.username || '用户'}
        </Button>
      </Dropdown>
    </div>
  );
}
