import { Card, Statistic, Row, Col } from 'antd';
import { ExperimentOutlined, UserOutlined, FileTextOutlined, CheckCircleOutlined } from '@ant-design/icons';
import { usePermission } from '../hooks/usePermission';
import { useAuthStore } from '../store/auth';

export default function Dashboard() {
  const { isAdmin } = usePermission();
  const user = useAuthStore((s) => s.user);

  const roleLabels: Record<string, string> = {
    super_admin: '超级管理员',
    lab_admin: '实验室负责人',
    member: '普通成员',
  };

  return (
    <>
      <div className="page-header">
        <h2>欢迎回来，{user?.username || '用户'}</h2>
        <p style={{ color: 'var(--color-text-secondary)' }}>
          当前角色：{roleLabels[user?.role_name || ''] || '—'}
        </p>
      </div>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="设备总数" prefix={<ExperimentOutlined />} value="—" />
          </Card>
        </Col>
        {isAdmin && (
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic title="用户总数" prefix={<UserOutlined />} value="—" />
            </Card>
          </Col>
        )}
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="我的借阅" prefix={<FileTextOutlined />} value="—" />
          </Card>
        </Col>
        {isAdmin && (
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic title="待审批" prefix={<CheckCircleOutlined />} value="—" />
            </Card>
          </Col>
        )}
      </Row>

      <Card title="快速入口" style={{ marginTop: 16 }}>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={8}>
            <Card size="small" hoverable onClick={() => window.location.href = '/equipments'}
              style={{ textAlign: 'center' }}>
              <ExperimentOutlined style={{ fontSize: 32, color: 'var(--color-primary)' }} />
              <p style={{ marginTop: 8 }}>设备大厅</p>
            </Card>
          </Col>
          <Col xs={24} sm={8}>
            <Card size="small" hoverable onClick={() => window.location.href = '/borrows/my'}
              style={{ textAlign: 'center' }}>
              <FileTextOutlined style={{ fontSize: 32, color: 'var(--color-primary)' }} />
              <p style={{ marginTop: 8 }}>我的借阅</p>
            </Card>
          </Col>
          {isAdmin && (
            <Col xs={24} sm={8}>
              <Card size="small" hoverable onClick={() => window.location.href = '/users'}
                style={{ textAlign: 'center' }}>
                <UserOutlined style={{ fontSize: 32, color: 'var(--color-primary)' }} />
                <p style={{ marginTop: 8 }}>用户管理</p>
              </Card>
            </Col>
          )}
        </Row>
      </Card>
    </>
  );
}
