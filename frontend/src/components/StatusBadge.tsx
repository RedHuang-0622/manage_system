import { Tag } from 'antd';

interface StatusBadgeProps {
  type: 'equipment' | 'borrow' | 'user';
  value: number | string;
}

export default function StatusBadge({ type, value }: StatusBadgeProps) {
  if (type === 'equipment') {
    const status = Number(value);
    return status === 1 ? <Tag color="green">上架</Tag> : <Tag color="red">下架</Tag>;
  }

  if (type === 'user') {
    const status = Number(value);
    return status === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>;
  }

  if (type === 'borrow') {
    const map: Record<string, string> = {
      '申请中': 'blue',
      '已借出': 'orange',
      '已归还': 'green',
      '被拒绝': 'red',
    };
    return <Tag color={map[String(value)] || 'default'}>{String(value)}</Tag>;
  }

  return <Tag>{String(value)}</Tag>;
}
