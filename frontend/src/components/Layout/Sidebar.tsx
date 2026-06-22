import { useNavigate, useLocation } from 'react-router-dom';
import { Menu } from 'antd';
import {
  DashboardOutlined,
  ExperimentOutlined,
  UserOutlined,
  FileTextOutlined,
} from '@ant-design/icons';
import { usePermission } from '../../hooks/usePermission';

interface SidebarProps {
  collapsed: boolean;
}

export default function Sidebar({ collapsed }: SidebarProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const { isEquipManager, canManageUsers, roleName } = usePermission();

  const path = location.pathname;
  const selectedKey = path.startsWith('/equipments')
    ? '/equipments'
    : path.startsWith('/users')
      ? '/users'
      : path.startsWith('/borrows')
        ? '/borrows'
        : '/';

  // viewer: read-only, cannot apply borrow
  const isViewer = roleName === 'viewer';

  const menuItems = [
    {
      key: '/',
      icon: <DashboardOutlined />,
      label: '仪表盘',
    },
    {
      key: '/equipments',
      icon: <ExperimentOutlined />,
      label: '设备大厅',
    },
    {
      key: 'borrows-group',
      icon: <FileTextOutlined />,
      label: '借阅管理',
      children: [
        { key: '/borrows/my', label: '我的借阅' },
        ...(isViewer
          ? []
          : [{ key: '/borrows/apply', label: '发起申请' }]
        ),
        ...(isEquipManager
          ? [
              { key: '/borrows/pending', label: '待审批' },
              { key: '/borrows/all', label: '全部记录' },
            ]
          : []),
      ],
    },
    ...(canManageUsers
      ? [
          {
            key: '/users',
            icon: <UserOutlined />,
            label: '用户管理',
          },
        ]
      : []),
  ];

  return (
    <>
      <div
        style={{
          height: 56,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: '#fff',
          fontWeight: 700,
          fontSize: collapsed ? 14 : 16,
          whiteSpace: 'nowrap',
          overflow: 'hidden',
        }}
      >
        {collapsed ? 'LAB' : '实验室管理系统'}
      </div>
      <Menu
        theme="dark"
        mode="inline"
        selectedKeys={[selectedKey]}
        items={menuItems}
        onClick={({ key }) => navigate(key)}
      />
    </>
  );
}
